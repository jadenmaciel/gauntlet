use std::fs;
use std::path::{Path, PathBuf};

use syn::spanned::Spanned;
use syn::visit::{self, Visit};
use syn::{
    Arm, Attribute, BinOp, Block, ExprBinary, ExprForLoop, ExprIf, ExprLoop, ExprWhile, File, Item,
};
use walkdir::{DirEntry, WalkDir};

use crate::model::FunctionComplexity;

pub fn scan_dir(root: &Path) -> Result<Vec<FunctionComplexity>, String> {
    let canonical_root = fs::canonicalize(root)
        .map_err(|error| format!("resolving scan root {}: {error}", root.display()))?;
    let mut functions = Vec::new();

    let walker = WalkDir::new(&canonical_root)
        .into_iter()
        .filter_entry(|entry| !skip_entry(entry));

    for entry in walker {
        let entry =
            entry.map_err(|error| format!("walking {}: {error}", canonical_root.display()))?;
        if !entry.file_type().is_file() {
            continue;
        }
        if entry.file_name() == "build.rs" {
            continue;
        }
        if entry.path().extension().and_then(|ext| ext.to_str()) != Some("rs") {
            continue;
        }

        let file_path = entry.path();
        let source = fs::read_to_string(file_path)
            .map_err(|error| format!("reading {}: {error}", file_path.display()))?;
        let syntax = syn::parse_file(&source)
            .map_err(|error| format!("parsing {}: {error}", file_path.display()))?;

        let relative = normalize_relative(file_path, &canonical_root);
        collect_file_functions(&syntax, &relative, &mut functions);
    }

    Ok(functions)
}

fn skip_entry(entry: &DirEntry) -> bool {
    let name = entry.file_name().to_string_lossy();
    entry.file_type().is_dir()
        && matches!(
            name.as_ref(),
            ".git" | ".gauntlet-tools" | "target" | "tests" | "benches" | "examples"
        )
}

fn normalize_relative(path: &Path, root: &Path) -> String {
    let relative: PathBuf = path.strip_prefix(root).unwrap_or(path).to_path_buf();
    relative.to_string_lossy().replace('\\', "/")
}

fn collect_file_functions(file: &File, relative_path: &str, output: &mut Vec<FunctionComplexity>) {
    collect_items(&file.items, relative_path, output);
}

fn collect_items(items: &[Item], relative_path: &str, output: &mut Vec<FunctionComplexity>) {
    for item in items {
        match item {
            Item::Fn(function) => {
                if should_skip_attrs(&function.attrs) {
                    continue;
                }
                output.push(FunctionComplexity {
                    file: relative_path.to_string(),
                    line: function.sig.fn_token.span().start().line,
                    func_name: function.sig.ident.to_string(),
                    complexity: block_complexity(&function.block),
                });
            }
            Item::Impl(implementation) => {
                if should_skip_attrs(&implementation.attrs) {
                    continue;
                }
                for impl_item in &implementation.items {
                    if let syn::ImplItem::Fn(method) = impl_item {
                        if should_skip_attrs(&method.attrs) {
                            continue;
                        }
                        output.push(FunctionComplexity {
                            file: relative_path.to_string(),
                            line: method.sig.fn_token.span().start().line,
                            func_name: method.sig.ident.to_string(),
                            complexity: block_complexity(&method.block),
                        });
                    }
                }
            }
            Item::Trait(trait_item) => {
                if should_skip_attrs(&trait_item.attrs) {
                    continue;
                }
                for item in &trait_item.items {
                    let syn::TraitItem::Fn(method) = item else {
                        continue;
                    };
                    if should_skip_attrs(&method.attrs) {
                        continue;
                    }
                    let Some(block) = &method.default else {
                        continue;
                    };
                    output.push(FunctionComplexity {
                        file: relative_path.to_string(),
                        line: method.sig.fn_token.span().start().line,
                        func_name: method.sig.ident.to_string(),
                        complexity: block_complexity(block),
                    });
                }
            }
            Item::Mod(module) => {
                if should_skip_attrs(&module.attrs) {
                    continue;
                }
                if let Some((_, module_items)) = &module.content {
                    collect_items(module_items, relative_path, output);
                }
            }
            _ => {}
        }
    }
}

fn should_skip_attrs(attrs: &[Attribute]) -> bool {
    attrs.iter().any(|attr| {
        if attr.path().is_ident("test") {
            return true;
        }
        if !attr.path().is_ident("cfg") {
            return false;
        }
        attr.meta
            .require_list()
            .map(|list| list.tokens.to_string().contains("test"))
            .unwrap_or(false)
    })
}

fn block_complexity(block: &Block) -> u64 {
    let mut visitor = DecisionVisitor { count: 1 };
    visitor.visit_block(block);
    visitor.count
}

struct DecisionVisitor {
    count: u64,
}

impl<'ast> Visit<'ast> for DecisionVisitor {
    fn visit_expr_if(&mut self, node: &'ast ExprIf) {
        self.count += 1;
        visit::visit_expr_if(self, node);
    }

    fn visit_expr_for_loop(&mut self, node: &'ast ExprForLoop) {
        self.count += 1;
        visit::visit_expr_for_loop(self, node);
    }

    fn visit_expr_while(&mut self, node: &'ast ExprWhile) {
        self.count += 1;
        visit::visit_expr_while(self, node);
    }

    fn visit_expr_loop(&mut self, node: &'ast ExprLoop) {
        self.count += 1;
        visit::visit_expr_loop(self, node);
    }

    fn visit_expr_binary(&mut self, node: &'ast ExprBinary) {
        if matches!(node.op, BinOp::And(_) | BinOp::Or(_)) {
            self.count += 1;
        }
        visit::visit_expr_binary(self, node);
    }

    fn visit_arm(&mut self, node: &'ast Arm) {
        if !matches!(node.pat, syn::Pat::Wild(_)) {
            self.count += 1;
        }
        visit::visit_arm(self, node);
    }
}

#[cfg(test)]
mod tests {
    use super::scan_dir;
    use std::fs;
    use std::path::PathBuf;
    use std::time::{SystemTime, UNIX_EPOCH};

    #[test]
    fn computes_per_function_complexity() {
        let dir = new_temp_dir("complexity");
        let src_dir = dir.join("src");
        fs::create_dir_all(&src_dir).expect("create src dir");

        let source = r#"
pub fn alpha(flag: bool) -> i32 {
    if flag { 1 } else { 0 }
}

impl Demo {
    pub fn beta(v: i32) -> i32 {
        match v {
            0 => 0,
            _ => 1,
        }
    }
}

#[cfg(test)]
mod tests {
    #[test]
    fn ignored() {}
}
"#;
        fs::write(src_dir.join("lib.rs"), source).expect("write source");

        let mut functions = scan_dir(&dir).expect("scan dir");
        functions.sort_by(|a, b| a.func_name.cmp(&b.func_name));

        assert_eq!(functions.len(), 2);
        assert_eq!(functions[0].func_name, "alpha");
        assert_eq!(functions[0].complexity, 2);
        assert_eq!(functions[1].func_name, "beta");
        assert_eq!(functions[1].complexity, 2);
    }

    #[test]
    fn includes_default_trait_methods_and_skips_non_product_targets() {
        let dir = new_temp_dir("targets");
        let src_dir = dir.join("src");
        fs::create_dir_all(&src_dir).expect("create src dir");
        fs::create_dir_all(dir.join("tests")).expect("create tests dir");
        fs::create_dir_all(dir.join("examples")).expect("create examples dir");

        fs::write(
            src_dir.join("lib.rs"),
            "pub trait Demo {\n    fn choose(&self, flag: bool) -> i32 {\n        if flag { 1 } else { 0 }\n    }\n}\n",
        )
        .expect("write source");
        fs::write(
            dir.join("tests/integration.rs"),
            "fn helper() -> bool { true }\n",
        )
        .expect("write integration test");
        fs::write(dir.join("examples/demo.rs"), "fn main() { if true {} }\n")
            .expect("write example");
        fs::write(dir.join("build.rs"), "fn main() { if true {} }\n").expect("write build script");

        let functions = scan_dir(&dir).expect("scan dir");

        assert_eq!(functions.len(), 1);
        assert_eq!(functions[0].func_name, "choose");
        assert_eq!(functions[0].complexity, 2);
    }

    fn new_temp_dir(prefix: &str) -> PathBuf {
        let suffix = SystemTime::now()
            .duration_since(UNIX_EPOCH)
            .expect("clock before epoch")
            .as_nanos();
        let dir = std::env::temp_dir().join(format!("crap4rs-{prefix}-{suffix}"));
        fs::create_dir_all(&dir).expect("create temp dir");
        dir
    }
}

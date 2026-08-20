import fs from "node:fs";
import path from "node:path";
import ts from "typescript";

import type { FunctionComplexity } from "./types.js";

const SOURCE_EXTENSIONS = new Set([".ts", ".tsx", ".mts", ".cts"]);
const SKIP_DIRECTORIES = new Set([".git", "node_modules", "dist", "coverage"]);

export function scanComplexity(dir: string): FunctionComplexity[] {
  const absDir = path.resolve(dir);
  const files = listSourceFiles(absDir);
  const program = ts.createProgram(files, {
    jsx: ts.JsxEmit.Preserve,
    noLib: true,
    noResolve: true,
    target: ts.ScriptTarget.Latest,
  });
  const functions: FunctionComplexity[] = [];

  for (const filePath of files) {
    const relativeFile = toPosix(path.relative(absDir, filePath));
    const sourceFile = program.getSourceFile(filePath);
    if (!sourceFile) {
      throw new Error(`failed to parse ${relativeFile}`);
    }
    const parseError = program.getSyntacticDiagnostics(sourceFile)[0];
    if (parseError) {
      const message = ts.flattenDiagnosticMessageText(parseError.messageText, "\n");
      const line =
        parseError.file && parseError.start !== undefined
          ? parseError.file.getLineAndCharacterOfPosition(parseError.start).line + 1
          : 1;
      throw new Error(`${relativeFile}:${line}: ${message}`);
    }
    functions.push(...complexityFromFile(sourceFile, relativeFile));
  }

  functions.sort((a, b) => {
    if (a.file !== b.file) {
      return a.file.localeCompare(b.file);
    }
    if (a.line !== b.line) {
      return a.line - b.line;
    }
    return a.func.localeCompare(b.func);
  });

  return functions;
}

function listSourceFiles(dir: string): string[] {
  const results: string[] = [];
  const stack: string[] = [dir];

  while (stack.length > 0) {
    const current = stack.pop()!;
    const entries = fs.readdirSync(current, { withFileTypes: true });

    for (const entry of entries) {
      const fullPath = path.join(current, entry.name);
      if (entry.isDirectory()) {
        if (!SKIP_DIRECTORIES.has(entry.name)) {
          stack.push(fullPath);
        }
        continue;
      }
      if (!entry.isFile()) {
        continue;
      }
      if (!SOURCE_EXTENSIONS.has(path.extname(entry.name)) || entry.name.endsWith(".d.ts")) {
        continue;
      }
      results.push(fullPath);
    }
  }

  return results;
}

function complexityFromFile(sourceFile: ts.SourceFile, relativeFile: string): FunctionComplexity[] {
  const results: FunctionComplexity[] = [];

  function visit(node: ts.Node): void {
    if (isFunctionLikeWithBody(node)) {
      const line = declarationLine(sourceFile, node);
      results.push({
        file: relativeFile,
        line,
        func: functionName(sourceFile, node),
        complexity: cyclomaticComplexity(node),
      });
    }
    ts.forEachChild(node, visit);
  }

  visit(sourceFile);
  return results;
}

type FunctionLikeWithBody =
  | ts.FunctionDeclaration
  | ts.MethodDeclaration
  | ts.GetAccessorDeclaration
  | ts.SetAccessorDeclaration
  | ts.ConstructorDeclaration
  | ts.FunctionExpression
  | ts.ArrowFunction;

function isFunctionLikeWithBody(node: ts.Node): node is FunctionLikeWithBody {
  return (
    (ts.isFunctionDeclaration(node) ||
      ts.isMethodDeclaration(node) ||
      ts.isGetAccessorDeclaration(node) ||
      ts.isSetAccessorDeclaration(node) ||
      ts.isConstructorDeclaration(node) ||
      ts.isFunctionExpression(node) ||
      ts.isArrowFunction(node)) &&
    node.body !== undefined
  );
}

function functionName(sourceFile: ts.SourceFile, node: FunctionLikeWithBody): string {
  if (ts.isFunctionDeclaration(node) && node.name) {
    return node.name.text;
  }
  if ((ts.isMethodDeclaration(node) || ts.isGetAccessorDeclaration(node) || ts.isSetAccessorDeclaration(node)) && node.name) {
    const methodName = propertyNameText(node.name);
    const className = node.parent && ts.isClassLike(node.parent) && node.parent.name ? node.parent.name.text : "";
    return className ? `${className}.${methodName}` : methodName;
  }
  if (ts.isConstructorDeclaration(node)) {
    if (node.parent && ts.isClassLike(node.parent) && node.parent.name) {
      return `${node.parent.name.text}.constructor`;
    }
    return "constructor";
  }
  if ((ts.isFunctionExpression(node) || ts.isArrowFunction(node)) && node.parent) {
    if (ts.isVariableDeclaration(node.parent) && ts.isIdentifier(node.parent.name)) {
      return node.parent.name.text;
    }
    if (ts.isPropertyAssignment(node.parent)) {
      return propertyNameText(node.parent.name);
    }
    if (ts.isBinaryExpression(node.parent) && node.parent.operatorToken.kind === ts.SyntaxKind.EqualsToken) {
      const name = assignmentTargetName(node.parent.left);
      if (name) {
        return name;
      }
    }
  }
  if ("name" in node && node.name && ts.isIdentifier(node.name)) {
    return node.name.text;
  }
  return `<anonymous@${lineOf(sourceFile, node)}>`;
}

function propertyNameText(name: ts.PropertyName): string {
  if (ts.isIdentifier(name) || ts.isPrivateIdentifier(name) || ts.isStringLiteral(name) || ts.isNumericLiteral(name)) {
    return name.text;
  }
  return name.getText();
}

function assignmentTargetName(node: ts.Expression): string | null {
  if (ts.isIdentifier(node)) {
    return node.text;
  }
  if (ts.isPropertyAccessExpression(node)) {
    return node.name.text;
  }
  if (ts.isElementAccessExpression(node)) {
    return node.argumentExpression?.getText() ?? null;
  }
  return null;
}

function lineOf(sourceFile: ts.SourceFile, node: ts.Node): number {
  return sourceFile.getLineAndCharacterOfPosition(node.getStart(sourceFile)).line + 1;
}

function declarationLine(sourceFile: ts.SourceFile, node: FunctionLikeWithBody): number {
  if ("name" in node && node.name) {
    return lineOf(sourceFile, node.name);
  }
  if (ts.isConstructorDeclaration(node)) {
    const keyword = node.getChildren(sourceFile).find((child) => child.kind === ts.SyntaxKind.ConstructorKeyword);
    if (keyword) {
      return lineOf(sourceFile, keyword);
    }
  }
  return lineOf(sourceFile, node);
}

function cyclomaticComplexity(node: FunctionLikeWithBody): number {
  let complexity = 1;
  const body = node.body;
  if (!body) {
    return complexity;
  }

  function visit(current: ts.Node): void {
    if (current !== node && ts.isFunctionLike(current)) {
      return;
    }

    if (
      ts.isIfStatement(current) ||
      ts.isForStatement(current) ||
      ts.isForInStatement(current) ||
      ts.isForOfStatement(current) ||
      ts.isWhileStatement(current) ||
      ts.isDoStatement(current) ||
      ts.isCatchClause(current) ||
      ts.isConditionalExpression(current)
    ) {
      complexity += 1;
    } else if (ts.isCaseClause(current) && current.expression) {
      complexity += 1;
    } else if (ts.isBinaryExpression(current) && isDecisionOperator(current.operatorToken.kind)) {
      complexity += 1;
    }

    ts.forEachChild(current, visit);
  }

  visit(body);
  return complexity;
}

function isDecisionOperator(kind: ts.SyntaxKind): boolean {
  return (
    kind === ts.SyntaxKind.BarBarToken ||
    kind === ts.SyntaxKind.AmpersandAmpersandToken ||
    kind === ts.SyntaxKind.QuestionQuestionToken ||
    kind === ts.SyntaxKind.BarBarEqualsToken ||
    kind === ts.SyntaxKind.AmpersandAmpersandEqualsToken ||
    kind === ts.SyntaxKind.QuestionQuestionEqualsToken
  );
}

function toPosix(value: string): string {
  return value.split(path.sep).join("/");
}

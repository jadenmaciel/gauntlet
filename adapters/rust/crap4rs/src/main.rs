use std::io;

fn main() {
    let args: Vec<String> = std::env::args().skip(1).collect();
    let code = crap4rs::run(&args, &mut io::stdout(), &mut io::stderr());
    std::process::exit(code);
}

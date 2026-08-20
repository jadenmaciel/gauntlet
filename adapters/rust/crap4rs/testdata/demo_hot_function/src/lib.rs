pub fn hot_path(input: i32) -> i32 {
    if input < 0 {
        -1
    } else if input == 0 {
        0
    } else if input % 2 == 0 {
        2
    } else if input % 3 == 0 {
        3
    } else if input % 5 == 0 {
        5
    } else if input % 7 == 0 {
        7
    } else {
        1
    }
}

#[cfg(test)]
mod tests {
    use super::hot_path;

    #[test]
    fn smoke_only() {
        assert_eq!(hot_path(2), 2);
    }

    #[test]
    fn broad_coverage() {
        assert_eq!(hot_path(-4), -1);
        assert_eq!(hot_path(0), 0);
        assert_eq!(hot_path(2), 2);
        assert_eq!(hot_path(9), 3);
        assert_eq!(hot_path(25), 5);
        assert_eq!(hot_path(49), 7);
        assert_eq!(hot_path(11), 1);
    }
}

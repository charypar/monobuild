#[derive(Clone, Copy, Debug, PartialEq, Eq, PartialOrd, Ord, Hash)]
pub enum Dependency {
    Weak,
    Strong,
}

// TODO bring some shared algorithms in here

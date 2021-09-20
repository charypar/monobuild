use std::{error::Error, fmt::Display, io};

pub type Commit = String;
pub type Command = Vec<String>;

pub enum Mode {
    Feature(String), // base branch, e.g. 'main'
    Main(String),    // base commit, e.g. 'HEAD^1'
}

#[derive(PartialEq, Debug)]
pub enum GitError {
    MergeBase(String, String), // base branch, error
    Diff(String),              // error
}
impl Error for GitError {}

impl Display for GitError {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            GitError::MergeBase(base_branch, error) => write!(
                f,
                "Cannot find merge base with branch {}: {}",
                base_branch, error
            ),
            GitError::Diff(error) => write!(f, "Finding changed files failed: {}", error),
        }
    }
}

pub struct Git<Executor>
where
    Executor: FnMut(Command) -> Result<String, String>,
{
    // Inversion of control for command execution to make Git pure
    // and easier to test
    executor: Executor,
}

impl<Executor> Git<Executor>
where
    Executor: FnMut(Command) -> Result<String, String>,
{
    pub fn new(executor: Executor) -> Self {
        Self { executor }
    }

    pub fn diff_base(&mut self, mode: Mode) -> Result<Commit, GitError> {
        match mode {
            Mode::Feature(base_branch) => (self.executor)(
                ["git", "merge-base", base_branch.as_ref(), "HEAD"]
                    .iter()
                    .map(|p| p.to_string())
                    .collect(),
            )
            .map(|base| base.trim_end().to_string())
            .map_err(|e| GitError::MergeBase(base_branch, e.to_string())),
            Mode::Main(base_commit) => Ok(base_commit.trim_end().to_string()),
        }
    }

    pub fn diff(&mut self, mode: Mode) -> Result<Vec<String>, GitError> {
        let base = self.diff_base(mode)?;

        (self.executor)(
            [
                "git",
                "diff",
                "--no-commit-id",
                "--name-only",
                "-r",
                base.as_ref(),
            ]
            .iter()
            .map(|p| p.to_string())
            .collect(),
        )
        .map(|files| {
            files
                .trim_end()
                .split("\n")
                .map(|f| f.to_string())
                .collect()
        })
        .map_err(|e| GitError::Diff(e.to_string()))
    }
}

#[cfg(test)]
mod test {
    mod diff_base {
        use std::io;

        use super::super::*;

        #[test]
        fn base_on_feature_branch() {
            let mut actual_command: Option<Command> = None;
            let expected_command = Some(vec![
                "git".into(),
                "merge-base".into(),
                "main".into(),
                "HEAD".into(),
            ]);

            let mock_exec = |cmd: Command| -> Result<String, String> {
                actual_command = Some(cmd);

                Ok("abc\n".to_string()) // check new line is trimmed
            };

            let mut git = Git::new(mock_exec);

            let actual = git.diff_base(Mode::Feature("main".to_string()));
            let expected = Ok("abc".to_string());

            assert_eq!(actual, expected);
            assert_eq!(actual_command, expected_command);
        }

        #[test]
        fn base_on_main_branch() {
            let mut actual_command: Option<Command> = None;
            let expected_command = None;

            let mock_exec = |cmd: Command| -> Result<String, String> {
                actual_command = Some(cmd);

                Ok("abc\n".to_string())
            };

            let mut git = Git::new(mock_exec);

            let actual = git.diff_base(Mode::Main("HEAD^1".to_string()));
            let expected = Ok("HEAD^1".to_string());

            assert_eq!(actual, expected);
            assert_eq!(actual_command, expected_command);
        }
    }

    mod diff {
        use std::io;

        use super::super::*;

        #[test]
        fn diff_on_feature_branch() {
            let mut actual_commands: Vec<Command> = vec![];
            let expected_command: Vec<String> = vec![
                "git".into(),
                "diff".into(),
                "--no-commit-id".into(),
                "--name-only".into(),
                "-r".into(),
                "main".into(),
            ];

            let mock_exec = |cmd: Command| -> Result<String, String> {
                actual_commands.push(cmd);

                if actual_commands.len() < 2 {
                    Ok("main\n".to_string())
                } else {
                    Ok("one\ntwo\nthree\n".to_string())
                }
            };

            let mut git = Git::new(mock_exec);

            let actual = git.diff(Mode::Feature("main".to_string()));
            let expected = Ok(vec![
                "one".to_string(),
                "two".to_string(),
                "three".to_string(),
            ]);

            assert_eq!(actual, expected);
            assert_eq!(actual_commands[1], expected_command);
        }

        #[test]
        fn diff_on_main_branch() {
            let mut actual_commands: Vec<Command> = vec![];
            let expected_command: Vec<String> = vec![
                "git".into(),
                "diff".into(),
                "--no-commit-id".into(),
                "--name-only".into(),
                "-r".into(),
                "HEAD^1".into(),
            ];

            let mock_exec = |cmd: Command| -> Result<String, String> {
                actual_commands.push(cmd);

                Ok("one\ntwo\nthree\n".to_string())
            };

            let mut git = Git::new(mock_exec);

            let actual = git.diff(Mode::Main("HEAD^1".to_string()));
            let expected = Ok(vec![
                "one".to_string(),
                "two".to_string(),
                "three".to_string(),
            ]);

            assert_eq!(actual, expected);
            assert_eq!(actual_commands[0], expected_command);
        }
    }
}

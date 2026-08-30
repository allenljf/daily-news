#!/usr/bin/env python3

import argparse
import fnmatch
import re
import subprocess
import sys
from dataclasses import dataclass
from pathlib import Path
from typing import Dict, Iterable, List, Optional, Sequence, Set, Tuple

import yaml


SUPPORTED_FILENAMES = {"pubspec.yaml", "analysis_options.yaml"}
SUPPORTED_SUFFIXES = {".dart"}
IGNORE_PATTERN = re.compile(r"guide-ignore:\s*([a-z0-9-]+(?:\s*,\s*[a-z0-9-]+)*)")
SELF_CHECK_SUCCESS = "Self-check passed"


@dataclass(frozen=True)
class Rule:
    rule_id: str
    owner: str
    check: str
    message: str
    pattern: Optional[str]
    applies_to: Tuple[str, ...]
    exclude: Tuple[str, ...]
    file_check: Optional[Dict[str, object]]


@dataclass(frozen=True)
class Violation:
    path: Path
    line: int
    rule_id: str
    message: str

    def render(self) -> str:
        return f"{self.path}:{self.line}: {self.rule_id}: {self.message}"


def parse_args(argv: Sequence[str]) -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Check Flutter guide rules that are safe to enforce mechanically.",
    )
    parser.add_argument(
        "--rules",
        default=str(Path(__file__).resolve().parent.parent / "rules.yaml"),
        help="Path to rules.yaml (default: %(default)s)",
    )

    target_group = parser.add_mutually_exclusive_group(required=True)
    target_group.add_argument("--all", action="store_true", help="Scan all supported files.")
    target_group.add_argument(
        "--staged",
        action="store_true",
        help="Scan staged supported files from git index.",
    )
    target_group.add_argument(
        "--diff",
        metavar="REVISION",
        help="Scan supported files changed from the given git revision or range.",
    )
    target_group.add_argument(
        "--files",
        nargs="+",
        metavar="PATH",
        help="Scan the given supported files.",
    )
    target_group.add_argument(
        "--self-check",
        action="store_true",
        help="Validate the current rules index and owner guide references.",
    )
    return parser.parse_args(argv)


def main(argv: Sequence[str]) -> int:
    args = parse_args(argv)
    try:
        rules_path = Path(args.rules).resolve()
        guide_root = rules_path.parent
        rules = load_rules(rules_path)

        if args.self_check:
            problems = validate_rules(rules, rules_path, guide_root)
            if problems:
                for problem in problems:
                    print(problem, file=sys.stderr)
                return 2
            print(f"{SELF_CHECK_SUCCESS}: {rules_path}")
            return 0

        candidate_files = resolve_candidate_files(args, guide_root.parent)
        violations = run_checks(rules, candidate_files, guide_root.parent)
        if violations:
            for violation in violations:
                print(violation.render())
            return 1
        return 0
    except RuleConfigurationError as error:
        print(str(error), file=sys.stderr)
        return 2
    except subprocess.CalledProcessError as error:
        stderr = error.stderr.strip() if error.stderr else ""
        message = stderr or str(error)
        print(message, file=sys.stderr)
        return 2


class RuleConfigurationError(Exception):
    pass


def load_rules(rules_path: Path) -> List[Rule]:
    try:
        with rules_path.open("r", encoding="utf-8") as handle:
            raw_rules = yaml.safe_load(handle)
    except FileNotFoundError as error:
        raise RuleConfigurationError(f"Rules file not found: {rules_path}") from error

    if not isinstance(raw_rules, list):
        raise RuleConfigurationError(f"Rules file must contain a top-level list: {rules_path}")

    rules: List[Rule] = []
    for entry in raw_rules:
        if not isinstance(entry, dict):
            raise RuleConfigurationError("Each rule entry must be a mapping.")
        rules.append(
            Rule(
                rule_id=required_string(entry, "id"),
                owner=required_string(entry, "owner"),
                check=required_string(entry, "check"),
                message=optional_string(entry, "message"),
                pattern=optional_string(entry, "pattern"),
                applies_to=tuple(string_list(entry.get("applies_to"))),
                exclude=tuple(string_list(entry.get("exclude"))),
                file_check=optional_mapping(entry, "file_check"),
            )
        )
    return rules


def required_string(entry: Dict[str, object], key: str) -> str:
    value = entry.get(key)
    if not isinstance(value, str) or not value.strip():
        raise RuleConfigurationError(f"Rule field '{key}' must be a non-empty string.")
    return value


def optional_string(entry: Dict[str, object], key: str) -> Optional[str]:
    value = entry.get(key)
    if value is None:
        return None
    if not isinstance(value, str):
        raise RuleConfigurationError(f"Rule field '{key}' must be a string when present.")
    return value


def optional_mapping(entry: Dict[str, object], key: str) -> Optional[Dict[str, object]]:
    value = entry.get(key)
    if value is None:
        return None
    if not isinstance(value, dict):
        raise RuleConfigurationError(f"Rule field '{key}' must be a mapping when present.")
    return value


def string_list(value: object) -> List[str]:
    if value is None:
        return []
    if not isinstance(value, list) or not all(isinstance(item, str) for item in value):
        raise RuleConfigurationError("Rule list fields must contain only strings.")
    return list(value)


def validate_rules(rules: Sequence[Rule], rules_path: Path, guide_root: Path) -> List[str]:
    problems: List[str] = []
    seen_rule_ids: Set[str] = set()

    for rule in rules:
        if rule.rule_id in seen_rule_ids:
            problems.append(f"Duplicate rule id: {rule.rule_id}")
        else:
            seen_rule_ids.add(rule.rule_id)

        owner_path = guide_root / "guides" / rule.owner
        if not owner_path.is_file():
            problems.append(f"Owner guide missing for {rule.rule_id}: {owner_path}")
            continue

        if not guide_declares_rule(owner_path, rule.rule_id):
            problems.append(
                f"Owner guide does not declare heading for {rule.rule_id}: {owner_path}"
            )

        if rule.check == "regex":
            if not rule.pattern:
                problems.append(f"Regex rule missing pattern: {rule.rule_id}")
            else:
                try:
                    re.compile(rule.pattern, re.MULTILINE)
                except re.error as error:
                    problems.append(f"Invalid regex for {rule.rule_id}: {error}")

            if not rule.applies_to:
                problems.append(f"Regex rule missing applies_to globs: {rule.rule_id}")
            if not rule.message:
                problems.append(f"Regex rule missing message: {rule.rule_id}")
        elif rule.check == "file":
            if not rule.file_check:
                problems.append(f"File rule missing file_check: {rule.rule_id}")
            if not rule.applies_to:
                problems.append(f"File rule missing applies_to globs: {rule.rule_id}")
            if not rule.message:
                problems.append(f"File rule missing message: {rule.rule_id}")
            problems.extend(validate_file_check(rule))
        elif rule.check != "manual":
            problems.append(f"Unsupported check type for {rule.rule_id}: {rule.check}")

        if rule.check in {"regex", "file"} and not supported_globs(rule.applies_to):
            problems.append(
                f"Rule {rule.rule_id} targets unsupported files; only .dart, pubspec.yaml, "
                "and analysis_options.yaml are allowed."
            )

    if not rules_path.is_file():
        problems.append(f"Rules file missing: {rules_path}")

    return problems


def supported_globs(patterns: Sequence[str]) -> bool:
    for pattern in patterns:
        if pattern.endswith(".dart") or pattern.endswith("pubspec.yaml") or pattern.endswith(
            "analysis_options.yaml"
        ):
            continue
        if "/pubspec.yaml" in pattern or "/analysis_options.yaml" in pattern:
            continue
        if pattern == "**/*.dart":
            continue
        return False
    return True


def validate_file_check(rule: Rule) -> List[str]:
    assert rule.file_check is not None
    problems: List[str] = []
    kind = rule.file_check.get("kind")
    if kind not in {"forbidden_regex", "required_regex"}:
        problems.append(
            f"Unsupported file_check.kind for {rule.rule_id}: {kind!r}. "
            "Use 'forbidden_regex' or 'required_regex'."
        )
        return problems
    pattern = rule.file_check.get("pattern")
    if not isinstance(pattern, str) or not pattern:
        problems.append(f"file_check.pattern must be a non-empty string for {rule.rule_id}")
        return problems
    try:
        re.compile(pattern, re.MULTILINE)
    except re.error as error:
        problems.append(f"Invalid file_check regex for {rule.rule_id}: {error}")
    return problems


def guide_declares_rule(owner_path: Path, rule_id: str) -> bool:
    heading_pattern = re.compile(rf"^### {re.escape(rule_id)} ·", re.MULTILINE)
    return bool(heading_pattern.search(owner_path.read_text(encoding="utf-8")))


def resolve_candidate_files(args: argparse.Namespace, repo_root: Path) -> List[Path]:
    if args.all:
        return discover_all_supported_files(repo_root)
    if args.staged:
        return git_changed_files(["diff", "--cached", "--name-only", "--diff-filter=ACMR"], repo_root)
    if args.diff:
        return git_changed_files(["diff", "--name-only", "--diff-filter=ACMR", args.diff], repo_root)
    if args.files:
        return normalize_and_filter_paths([Path(path) for path in args.files], repo_root)
    raise RuleConfigurationError("One scan target must be selected.")


def discover_all_supported_files(repo_root: Path) -> List[Path]:
    discovered: List[Path] = []
    for path in repo_root.rglob("*"):
        if path.is_file() and is_supported_file(path):
            discovered.append(path.resolve())
    return sorted(set(discovered))


def git_changed_files(command: Sequence[str], repo_root: Path) -> List[Path]:
    completed = subprocess.run(
        ["git", *command],
        cwd=repo_root,
        text=True,
        capture_output=True,
        check=True,
    )
    paths = [repo_root / line for line in completed.stdout.splitlines() if line.strip()]
    return normalize_and_filter_paths(paths, repo_root)


def normalize_and_filter_paths(paths: Iterable[Path], repo_root: Path) -> List[Path]:
    normalized: List[Path] = []
    for path in paths:
        resolved = path if path.is_absolute() else (repo_root / path)
        resolved = resolved.resolve()
        if resolved.is_file() and is_supported_file(resolved):
            normalized.append(resolved)
    return sorted(set(normalized))


def is_supported_file(path: Path) -> bool:
    return path.suffix in SUPPORTED_SUFFIXES or path.name in SUPPORTED_FILENAMES


def run_checks(rules: Sequence[Rule], files: Sequence[Path], repo_root: Path) -> List[Violation]:
    violations: List[Violation] = []
    for path in files:
        text = path.read_text(encoding="utf-8")
        ignored_lines = collect_ignored_lines(text.splitlines())
        for rule in rules:
            if rule.check == "manual":
                continue
            if not path_matches_rule(path, rule, repo_root):
                continue
            if rule.check == "regex":
                violations.extend(check_regex_rule(rule, path, text, ignored_lines))
            elif rule.check == "file":
                violations.extend(check_file_rule(rule, path, text, ignored_lines))
    return sorted(violations, key=lambda item: (str(item.path), item.line, item.rule_id))


def collect_ignored_lines(lines: Sequence[str]) -> Dict[str, Set[int]]:
    ignored: Dict[str, Set[int]] = {}
    for index, line in enumerate(lines, start=1):
        match = IGNORE_PATTERN.search(line)
        if not match:
            continue
        target_line = index
        if line.strip().startswith("//"):
            target_line = next_nonempty_line(lines, index)
        for rule_id in [item.strip() for item in match.group(1).split(",")]:
            ignored.setdefault(rule_id, set()).update({index, target_line})
    return ignored


def next_nonempty_line(lines: Sequence[str], start_line: int) -> int:
    for line_number in range(start_line + 1, len(lines) + 1):
        if lines[line_number - 1].strip():
            return line_number
    return start_line


def path_matches_rule(path: Path, rule: Rule, repo_root: Path) -> bool:
    candidates = match_candidates(path, repo_root)
    if rule.applies_to and not any(
        any(fnmatch.fnmatchcase(candidate, pattern) for candidate in candidates)
        for pattern in rule.applies_to
    ):
        return False
    if rule.exclude and any(
        any(fnmatch.fnmatchcase(candidate, pattern) for candidate in candidates)
        for pattern in rule.exclude
    ):
        return False
    return True


def match_candidates(path: Path, repo_root: Path) -> List[str]:
    candidates = [path.as_posix(), path.name]
    try:
        candidates.append(path.relative_to(repo_root).as_posix())
    except ValueError:
        pass
    return candidates


def check_regex_rule(
    rule: Rule,
    path: Path,
    text: str,
    ignored_lines: Dict[str, Set[int]],
) -> List[Violation]:
    assert rule.pattern is not None
    compiled = re.compile(rule.pattern, re.MULTILINE)
    seen_lines: Set[int] = set()
    violations: List[Violation] = []

    for match in compiled.finditer(text):
        line = text.count("\n", 0, match.start()) + 1
        if line in seen_lines:
            continue
        if line in ignored_lines.get(rule.rule_id, set()):
            continue
        seen_lines.add(line)
        violations.append(
            Violation(
                path=path,
                line=line,
                rule_id=rule.rule_id,
                message=rule.message,
            )
        )
    return violations


def check_file_rule(
    rule: Rule,
    path: Path,
    text: str,
    ignored_lines: Dict[str, Set[int]],
) -> List[Violation]:
    assert rule.file_check is not None
    kind = rule.file_check["kind"]
    pattern = str(rule.file_check["pattern"])
    compiled = re.compile(pattern, re.MULTILINE)
    ignored_for_rule = ignored_lines.get(rule.rule_id, set())

    if kind == "forbidden_regex":
        violations: List[Violation] = []
        for match in compiled.finditer(text):
            line = text.count("\n", 0, match.start()) + 1
            if line in ignored_for_rule:
                continue
            violations.append(
                Violation(path=path, line=line, rule_id=rule.rule_id, message=rule.message)
            )
        return violations

    if compiled.search(text):
        return []
    if 1 in ignored_for_rule:
        return []
    return [Violation(path=path, line=1, rule_id=rule.rule_id, message=rule.message)]


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))

import ast
import re
import subprocess
import sys
import tempfile
import textwrap
import unittest
from pathlib import Path
from typing import Dict, List, Union


TOOLS_DIR = Path(__file__).resolve().parent
CHECKER_PATH = TOOLS_DIR / "check-rules.py"
RULES_PATH = TOOLS_DIR.parent / "rules.yaml"


Violation = Dict[str, Union[str, int]]


def parse_violations(output: str) -> List[Violation]:
    violations: List[Violation] = []
    for line in output.splitlines():
        match = re.match(r"^(.*?):(\d+): ([a-z0-9-]+): (.+)$", line)
        if not match:
            continue
        path_text, line_number, rule_id, message = match.groups()
        violations.append(
            {
                "path": path_text,
                "line": int(line_number),
                "rule_id": rule_id,
                "message": message,
            }
        )
    return violations


class CheckRulesCliTest(unittest.TestCase):
    maxDiff = None

    def run_checker(
        self,
        *paths: Path,
        rules_path: Path = RULES_PATH,
    ) -> subprocess.CompletedProcess[str]:
        command = [
            sys.executable,
            str(CHECKER_PATH),
            "--rules",
            str(rules_path),
            "--files",
            *[str(path) for path in paths],
        ]
        return subprocess.run(
            command,
            cwd=TOOLS_DIR.parent.parent,
            text=True,
            capture_output=True,
            check=False,
        )

    def write_fixture(self, root: Path, relative_path: str, contents: str) -> Path:
        target = root / relative_path
        target.parent.mkdir(parents=True, exist_ok=True)
        target.write_text(textwrap.dedent(contents).lstrip(), encoding="utf-8")
        return target

    def assert_single_violation(
        self,
        completed: subprocess.CompletedProcess[str],
        *,
        rule_id: str,
        source_line: int,
    ) -> None:
        self.assertEqual(
            completed.returncode,
            1,
            msg=f"stdout:\n{completed.stdout}\n\nstderr:\n{completed.stderr}",
        )
        violations = parse_violations(completed.stdout)
        self.assertEqual(len(violations), 1, completed.stdout)
        self.assertEqual(violations[0]["rule_id"], rule_id)
        self.assertEqual(violations[0]["line"], source_line)

    def test_flags_dio_usage_in_widget_file(self) -> None:
        with tempfile.TemporaryDirectory() as temp_dir:
            fixture = self.write_fixture(
                Path(temp_dir),
                "lib/features/news/presentation/news_page.dart",
                """
                import 'package:dio/dio.dart';
                import 'package:flutter/widgets.dart';

                class NewsPage extends StatelessWidget {
                  final Dio dio = Dio();

                  @override
                  Widget build(BuildContext context) {
                    return const SizedBox();
                  }
                }
                """,
            )

            completed = self.run_checker(fixture)

            self.assert_single_violation(
                completed,
                rule_id="dio-outside-remote-service",
                source_line=5,
            )

    def test_flags_repository_usage_in_widget_file(self) -> None:
        with tempfile.TemporaryDirectory() as temp_dir:
            fixture = self.write_fixture(
                Path(temp_dir),
                "lib/features/news/presentation/news_screen.dart",
                """
                import 'package:flutter/widgets.dart';

                class NewsRepository {}

                class NewsScreen extends StatelessWidget {
                  final NewsRepository repository;

                  const NewsScreen(this.repository, {super.key});

                  @override
                  Widget build(BuildContext context) {
                    return const SizedBox();
                  }
                }
                """,
            )

            completed = self.run_checker(fixture)

            self.assert_single_violation(
                completed,
                rule_id="widget-reads-repository",
                source_line=6,
            )

    def test_flags_hardcoded_secret_in_dart_file(self) -> None:
        with tempfile.TemporaryDirectory() as temp_dir:
            fixture = self.write_fixture(
                Path(temp_dir),
                "lib/core/security/api_keys.dart",
                """
                const geminiApiKey = 'AIzaSyARealSecretShouldNotBeHere';
                """,
            )

            completed = self.run_checker(fixture)

            self.assert_single_violation(
                completed,
                rule_id="secret-in-client-code",
                source_line=1,
            )

    def test_allows_guide_ignore_for_specific_rule(self) -> None:
        with tempfile.TemporaryDirectory() as temp_dir:
            fixture = self.write_fixture(
                Path(temp_dir),
                "lib/features/news/presentation/news_page.dart",
                """
                import 'package:dio/dio.dart';
                import 'package:flutter/widgets.dart';

                class NewsPage extends StatelessWidget {
                  // guide-ignore: dio-outside-remote-service
                  final Dio dio = Dio();

                  @override
                  Widget build(BuildContext context) {
                    return const SizedBox();
                  }
                }
                """,
            )

            completed = self.run_checker(fixture)

            self.assertEqual(
                completed.returncode,
                0,
                msg=f"stdout:\n{completed.stdout}\n\nstderr:\n{completed.stderr}",
            )
            self.assertEqual(parse_violations(completed.stdout), [])

    def test_allows_guide_ignore_for_custom_forbidden_file_rule(self) -> None:
        with tempfile.TemporaryDirectory() as temp_dir:
            temp_root = Path(temp_dir)
            fixture = self.write_fixture(
                temp_root,
                "lib/features/news/presentation/example.dart",
                """
                // guide-ignore: exact-bar-line
                bar
                """,
            )
            rules_path = self.write_fixture(
                temp_root,
                "rules.yaml",
                """
                - id: exact-bar-line
                  owner: 00-principles.md
                  check: file
                  message: exact bar forbidden
                  applies_to:
                    - "**/*.dart"
                  file_check:
                    kind: forbidden_regex
                    pattern: "^bar$"
                """,
            )
            self.write_fixture(
                temp_root,
                "guides/00-principles.md",
                """
                ## exact-bar-line
                """,
            )

            completed = self.run_checker(fixture, rules_path=rules_path)

            self.assertEqual(
                completed.returncode,
                0,
                msg=f"stdout:\n{completed.stdout}\n\nstderr:\n{completed.stderr}",
            )
            self.assertEqual(parse_violations(completed.stdout), [])

    def test_only_checks_supported_file_types(self) -> None:
        with tempfile.TemporaryDirectory() as temp_dir:
            fixture = self.write_fixture(
                Path(temp_dir),
                "README.md",
                """
                const geminiApiKey = 'AIzaSyStillIgnoredBecauseWrongFileType';
                """,
            )

            completed = self.run_checker(fixture)

            self.assertEqual(
                completed.returncode,
                0,
                msg=f"stdout:\n{completed.stdout}\n\nstderr:\n{completed.stderr}",
            )
            self.assertEqual(parse_violations(completed.stdout), [])

    def test_self_check_uses_current_rule_index(self) -> None:
        completed = subprocess.run(
            [
                sys.executable,
                str(CHECKER_PATH),
                "--rules",
                str(RULES_PATH),
                "--self-check",
            ],
            cwd=TOOLS_DIR.parent.parent,
            text=True,
            capture_output=True,
            check=False,
        )

        self.assertEqual(
            completed.returncode,
            0,
            msg=f"stdout:\n{completed.stdout}\n\nstderr:\n{completed.stderr}",
        )
        self.assertIn("Self-check passed", completed.stdout)

    def test_rule_fixture_remains_single_source_of_truth(self) -> None:
        data = ast.literal_eval(
            repr(
                parse_violations(
                    "lib/example.dart:7: dio-outside-remote-service: bad\n"
                )
            )
        )
        self.assertEqual(
            data,
            [
                {
                    "path": "lib/example.dart",
                    "line": 7,
                    "rule_id": "dio-outside-remote-service",
                    "message": "bad",
                }
            ],
        )


if __name__ == "__main__":
    unittest.main()

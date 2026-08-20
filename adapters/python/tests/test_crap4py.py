from __future__ import annotations

import ast
import io
import json
import tempfile
import unittest
from pathlib import Path

from adapters.python import crap4py


class Crap4PyTests(unittest.TestCase):
    def test_golden_vectors_match_formula(self) -> None:
        vectors = [
            (1, 0.0, 2.0),
            (1, 100.0, 1.0),
            (2, 0.0, 6.0),
            (3, 50.0, 4.125),
            (5, 0.0, 30.0),
            (5, 50.0, 8.125),
            (5, 100.0, 5.0),
            (8, 100.0, 8.0),
            (10, 0.0, 110.0),
            (10, 80.0, 10.8),
            (10, 100.0, 10.0),
            (15, 100.0, 15.0),
            (20, 90.0, 20.4),
        ]
        epsilon = 1e-9

        for complexity, coverage_percent, expected in vectors:
            actual = crap4py.compute_crap(complexity, coverage_percent)
            self.assertLess(abs(actual - expected), epsilon)

    def test_unmatched_function_defaults_to_zero_coverage_and_stays_scored(self) -> None:
        functions = [
            crap4py.FunctionComplexity(file="pkg/mod.py", line=10, func="covered", complexity=1),
            crap4py.FunctionComplexity(file="pkg/mod.py", line=20, func="missing", complexity=2),
        ]
        coverage = {
            crap4py.build_key("pkg/mod.py", 10, "covered"): crap4py.CoverageEntry(
                coverage=100.0
            )
        }

        report = crap4py.evaluate(functions, coverage, ceiling=10.0)

        self.assertEqual(report.summary.total, 2)
        covered = next(function for function in report.functions if function.func == "covered")
        self.assertEqual(covered.coverage, 100.0)
        self.assertAlmostEqual(covered.crap, 1.0)
        missing = next(function for function in report.functions if function.func == "missing")
        self.assertFalse(missing.matched)
        self.assertEqual(missing.coverage, 0.0)
        self.assertAlmostEqual(missing.crap, 6.0)

    def test_ast_complexity_counts_decisions_without_counting_lambda_declarations(self) -> None:
        plain_lambda = ast.parse("def f():\n    return lambda value: value\n").body[0]
        guarded_wildcard = ast.parse(
            "def f(value):\n"
            "    match value:\n"
            "        case _ if value > 0:\n"
            "            return value\n"
            "    return 0\n"
        ).body[0]

        self.assertEqual(crap4py.function_complexity(plain_lambda), 1)
        self.assertEqual(crap4py.function_complexity(guarded_wildcard), 2)

    def test_coverage_xml_joins_class_lines_by_file_line_and_function(self) -> None:
        with tempfile.TemporaryDirectory() as tmp_dir_name:
            tmp_dir = Path(tmp_dir_name)
            (tmp_dir / "sample.py").write_text(
                "def covered(x):\n"
                "    if x > 0:\n"
                "        return x\n"
                "    return -x\n"
                "\n"
                "def uncovered(y):\n"
                "    if y > 1:\n"
                "        return y\n"
                "    return 0\n",
                encoding="utf-8",
            )
            coverage_path = tmp_dir / "coverage.xml"
            coverage_path.write_text(
                '<?xml version="1.0" ?>\n'
                "<coverage>\n"
                "  <sources>\n"
                f"    <source>{tmp_dir}</source>\n"
                "  </sources>\n"
                "  <packages>\n"
                '    <package name=".">\n'
                "      <classes>\n"
                '        <class name="sample.py" filename="sample.py">\n'
                "          <methods/>\n"
                "          <lines>\n"
                '            <line number="1" hits="1"/>\n'
                '            <line number="2" hits="1"/>\n'
                '            <line number="3" hits="1"/>\n'
                '            <line number="4" hits="1"/>\n'
                '            <line number="6" hits="1"/>\n'
                '            <line number="7" hits="0"/>\n'
                '            <line number="8" hits="0"/>\n'
                '            <line number="9" hits="0"/>\n'
                "          </lines>\n"
                "        </class>\n"
                "      </classes>\n"
                "    </package>\n"
                "  </packages>\n"
                "</coverage>\n",
                encoding="utf-8",
            )
            functions = crap4py.scan_dir(str(tmp_dir))

            coverage = crap4py.load_coverage_xml(
                str(coverage_path), str(tmp_dir), functions
            )

            self.assertEqual(
                coverage[crap4py.build_key("sample.py", 1, "covered")].coverage,
                100.0,
            )
            self.assertEqual(
                coverage[crap4py.build_key("sample.py", 6, "uncovered")].coverage,
                0.0,
            )

    def test_run_reads_threshold_and_emits_json_shape(self) -> None:
        with tempfile.TemporaryDirectory() as tmp_dir_name:
            tmp_dir = Path(tmp_dir_name)
            source_path = tmp_dir / "sample.py"
            source_path.write_text(
                "def covered(x):\n"
                "    if x > 0:\n"
                "        return x\n"
                "    return -x\n"
                "\n"
                "def uncovered(y):\n"
                "    if y > 1:\n"
                "        return y\n"
                "    return 0\n",
                encoding="utf-8",
            )

            coverage_path = tmp_dir / "coverage.xml"
            coverage_path.write_text(
                '<?xml version="1.0" ?>\n'
                '<coverage>\n'
                '  <packages>\n'
                '    <package name=".">\n'
                '      <classes>\n'
                '        <class name="sample.py" filename="sample.py">\n'
                "          <methods/>\n"
                "          <lines>\n"
                '            <line number="1" hits="1"/>\n'
                '            <line number="2" hits="1"/>\n'
                '            <line number="3" hits="1"/>\n'
                '            <line number="4" hits="1"/>\n'
                '            <line number="6" hits="0"/>\n'
                '            <line number="7" hits="0"/>\n'
                '            <line number="8" hits="0"/>\n'
                '            <line number="9" hits="0"/>\n'
                "          </lines>\n"
                "        </class>\n"
                "      </classes>\n"
                "    </package>\n"
                "  </packages>\n"
                "</coverage>\n",
                encoding="utf-8",
            )

            thresholds_dir = tmp_dir / ".gauntlet"
            thresholds_dir.mkdir(parents=True)
            thresholds_path = thresholds_dir / "thresholds.yml"
            thresholds_path.write_text(
                "metrics:\n"
                "  crap_ceiling:\n"
                "    direction: max\n"
                "    value: 3\n",
                encoding="utf-8",
            )

            stdout = io.StringIO()
            stderr = io.StringIO()
            code = crap4py.run(
                [
                    "--dir",
                    str(tmp_dir),
                    "--coverage",
                    str(coverage_path),
                    "--thresholds",
                    str(thresholds_path),
                    "--format",
                    "json",
                ],
                stdout,
                stderr,
            )

            self.assertEqual(code, 1)
            self.assertEqual(stderr.getvalue(), "")
            output = json.loads(stdout.getvalue())
            self.assertEqual(set(output.keys()), {"ceiling", "functions", "summary"})
            self.assertEqual(output["ceiling"], 3.0)
            self.assertEqual(set(output["summary"].keys()), {"total", "failing", "max_crap"})
            self.assertGreaterEqual(output["summary"]["failing"], 1)
            by_name = {function["func"]: function for function in output["functions"]}
            self.assertEqual(by_name["covered"]["coverage"], 100.0)
            self.assertEqual(by_name["uncovered"]["coverage"], 0.0)

            for function in output["functions"]:
                self.assertEqual(
                    set(function.keys()),
                    {"file", "line", "func", "complexity", "coverage", "crap", "pass"},
                )


if __name__ == "__main__":
    unittest.main()

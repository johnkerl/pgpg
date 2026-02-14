/**
 * Unit tests for parser codegen: ensure hint 'type' is applied for with_adopted_grandchildren.
 */
import fs from "fs";
import os from "os";
import path from "path";
import { describe, it } from "node:test";
import assert from "node:assert";
import { fileURLToPath } from "url";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const codegenDir = path.join(__dirname, "..", "codegen");

async function runParsegenCode(outputPath, jsonPath, className, prefix = "") {
  const scriptPath = path.join(codegenDir, "parsegen_code.js");
  const args = [
    "-o", outputPath,
    "-c", className,
    "--prefix", prefix,
    jsonPath,
  ];
  const { spawn } = await import("node:child_process");
  const proc = spawn("node", [scriptPath, ...args], {
    stdio: "pipe",
    cwd: path.join(__dirname, ".."),
  });
  return new Promise((resolve) => {
    let err = "";
    proc.stderr?.on("data", (c) => (err += c));
    proc.on("close", (code) => resolve({ code, err }));
  });
}

describe("parsegen_code", () => {
  it("generated parser applies hint type in with_adopted_grandchildren branch", async () => {
    const tmp = os.tmpdir();
    const jsonPath = path.join(tmp, "pgpg_js_test_parsegen_" + Date.now() + ".json");
    const outPath = path.join(tmp, "pgpg_js_test_array_parser_" + Date.now() + ".js");
    const tables = {
      start_symbol: "Root",
      actions: { "0": { EOF: { type: "accept" } } },
      gotos: {},
      productions: [
        {
          lhs: "Root",
          rhs: [
            { name: "lbracket", terminal: true },
            { name: "Elements", terminal: false },
            { name: "rbracket", terminal: true },
          ],
          hint: {
            parent_literal: "[]",
            with_adopted_grandchildren: [1],
            type: "array",
          },
        },
      ],
      hint_mode: "hints",
    };
    fs.writeFileSync(jsonPath, JSON.stringify(tables), "utf-8");
    try {
      const { code } = await runParsegenCode(outPath, jsonPath, "ArrayParser", "");
      assert.strictEqual(code, 0, "parsegen_code should exit 0");
      const codeStr = fs.readFileSync(outPath, "utf-8");
      assert.ok(
        codeStr.includes("nodeType = prod.nodeType || parentType"),
        "generated code should apply hint type in with_adopted_grandchildren branch",
      );
      assert.ok(
        codeStr.includes('nodeType: "array"'),
        "generated production table should include node_type for the hint",
      );
      assert.ok(
        codeStr.includes("newASTNode(parentToken, nodeType, newChildren)"),
        "generated code should build node with nodeType in with_adopted_grandchildren branch",
      );
    } finally {
      try {
        fs.unlinkSync(jsonPath);
        fs.unlinkSync(outPath);
      } catch (_) {}
    }
  });
});

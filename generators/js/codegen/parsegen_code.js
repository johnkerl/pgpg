#!/usr/bin/env node
/**
 * Generate a JavaScript LR(1) parser from JSON tables (output of parsegen-tables).
 * Reads the same *-parse.json files as the Go parsegen-code.
 * Usage: node parsegen_code.js -o parsers/json_parser.js -c JSONParser [--prefix pgpg_] json-parse.json
 */
import fs from "fs";
import path from "path";
import { fileURLToPath } from "url";

const __dirname = path.dirname(fileURLToPath(import.meta.url));

function loadTables(filePath) {
  return JSON.parse(fs.readFileSync(filePath, "utf-8"));
}

function buildActions(raw) {
  const actions = raw.actions ?? {};
  return Object.keys(actions)
    .sort((a, b) => Number(a) - Number(b))
    .map((stateStr) => {
      const state = Number(stateStr);
      const terms = actions[stateStr];
      const entries = Object.keys(terms)
        .sort()
        .map((term) => {
          const act = terms[term];
          const kind =
            act.type === "shift"
              ? "ActionKind.SHIFT"
              : act.type === "reduce"
                ? "ActionKind.REDUCE"
                : "ActionKind.ACCEPT";
          return {
            terminalLiteral: JSON.stringify(term),
            kindLiteral: kind,
            target: act.target ?? 0,
          };
        });
      return { state, entries };
    });
}

function buildGotos(raw) {
  const gotos = raw.gotos ?? {};
  return Object.keys(gotos)
    .sort((a, b) => Number(a) - Number(b))
    .map((stateStr) => {
      const state = Number(stateStr);
      const entries = Object.keys(gotos[stateStr])
        .sort()
        .map((nonterm) => ({
          nontermLiteral: JSON.stringify(nonterm),
          target: gotos[stateStr][nonterm],
        }));
      return { state, entries };
    });
}

function listOrEmpty(v) {
  return v ?? [];
}

function buildProductions(raw) {
  const productions = raw.productions ?? [];
  return productions.map((prod) => {
    const rhs = prod.rhs ?? [];
    const rhsCount = rhs.length;
    const hint = prod.hint;
    const info = {
      lhsLiteral: JSON.stringify(prod.lhs),
      rhsCount,
      hasHint: false,
      hasPassthrough: false,
      hasParentLiteral: false,
      hasWithAppendedChildren: false,
      hasWithPrependedChildren: false,
      hasWithAdoptedGrandchildren: false,
      parentIndex: 0,
      passthroughIndex: 0,
      parentLiteral: "",
      childIndices: [],
      withAppendedChildren: [],
      withPrependedChildren: [],
      withAdoptedGrandchildren: [],
      nodeType: "",
    };
    if (hint != null) {
      if (hint["pass-through"] != null) {
        info.hasPassthrough = true;
        info.passthroughIndex = hint["pass-through"];
      } else {
        info.hasHint = true;
        if (hint["parent_literal"] != null) {
          info.hasParentLiteral = true;
          info.parentLiteral = hint["parent_literal"];
        } else {
          info.parentIndex = hint.parent ?? 0;
        }
        info.childIndices = listOrEmpty(hint.children);
        info.withAppendedChildren = listOrEmpty(hint.with_appended_children);
        info.hasWithAppendedChildren = info.withAppendedChildren.length > 0;
        info.withPrependedChildren = listOrEmpty(hint.with_prepended_children);
        info.hasWithPrependedChildren = info.withPrependedChildren.length > 0;
        info.withAdoptedGrandchildren = listOrEmpty(hint.with_adopted_grandchildren);
        info.hasWithAdoptedGrandchildren = info.withAdoptedGrandchildren.length > 0;
        info.nodeType = hint.type ?? "";
      }
    }
    return info;
  });
}

function serializeProductions(productions) {
  return productions
    .map(
      (p) =>
        `  { lhs: ${p.lhsLiteral}, rhsCount: ${p.rhsCount}, hasHint: ${p.hasHint}, hasPassthrough: ${p.hasPassthrough}, hasParentLiteral: ${p.hasParentLiteral}, hasWithAppendedChildren: ${p.hasWithAppendedChildren}, hasWithPrependedChildren: ${p.hasWithPrependedChildren}, hasWithAdoptedGrandchildren: ${p.hasWithAdoptedGrandchildren}, parentIndex: ${p.parentIndex}, passthroughIndex: ${p.passthroughIndex}, parentLiteral: ${JSON.stringify(p.parentLiteral)}, childIndices: ${JSON.stringify(p.childIndices)}, withAppendedChildren: ${JSON.stringify(p.withAppendedChildren)}, withPrependedChildren: ${JSON.stringify(p.withPrependedChildren)}, withAdoptedGrandchildren: ${JSON.stringify(p.withAdoptedGrandchildren)}, nodeType: ${JSON.stringify(p.nodeType)} }`,
    )
    .join(",\n");
}

function serializeActions(actions) {
  const lines = [];
  for (const a of actions) {
    const entries = a.entries
      .map(
        (e) =>
          `    ${e.terminalLiteral}: new Action(${e.kindLiteral}, ${e.target})`,
      )
      .join(",\n");
    lines.push(`  ${a.state}: {\n${entries}\n  }`);
  }
  return lines.join(",\n");
}

function serializeGotos(gotos) {
  const lines = [];
  for (const g of gotos) {
    const entries = g.entries
      .map((e) => `    ${e.nontermLiteral}: ${e.target}`)
      .join(",\n");
    lines.push(`  ${g.state}: {\n${entries}\n  }`);
  }
  return lines.join(",\n");
}

function renderParser(ctx) {
  const { className, actions, gotos, productions, hintMode } = ctx;
  const productionsStr = serializeProductions(productions);
  const actionsStr = serializeActions(actions);
  const gotosStr = serializeGotos(gotos);
  const hasHints = hintMode === "hints";

  const reduceBody = hasHints
    ? `
    if (!useFullTree && prod.hasPassthrough) {
      return rhsNodes[prod.passthroughIndex];
    }
    if (!useFullTree && prod.hasWithAppendedChildren) {
      let parent = null;
      let parentToken = null;
      let parentType = null;
      if (prod.hasParentLiteral) {
        parentToken = newToken(prod.parentLiteral, prod.parentLiteral);
        parentType = prod.parentLiteral;
      } else {
        parent = rhsNodes[prod.parentIndex];
        parentToken = parent.token;
        parentType = parent.type;
      }
      const nodeType = prod.nodeType || parentType;
      const newChildren = [];
      if (parent != null && parent.children?.length) newChildren.push(...parent.children);
      for (const ci of prod.withAppendedChildren) newChildren.push(rhsNodes[ci]);
      return newASTNode(parentToken, nodeType, newChildren);
    }
    if (!useFullTree && prod.hasWithPrependedChildren) {
      let parent = null;
      let parentToken = null;
      let parentType = null;
      if (prod.hasParentLiteral) {
        parentToken = newToken(prod.parentLiteral, prod.parentLiteral);
        parentType = prod.parentLiteral;
      } else {
        parent = rhsNodes[prod.parentIndex];
        parentToken = parent.token;
        parentType = parent.type;
      }
      const nodeType = prod.nodeType || parentType;
      const newChildren = [];
      for (const ci of prod.withPrependedChildren) newChildren.push(rhsNodes[ci]);
      if (parent != null && parent.children?.length) newChildren.push(...parent.children);
      return newASTNode(parentToken, nodeType, newChildren);
    }
    if (!useFullTree && prod.hasWithAdoptedGrandchildren) {
      let parent = null;
      let parentToken = null;
      let parentType = null;
      if (prod.hasParentLiteral) {
        parentToken = newToken(prod.parentLiteral, prod.parentLiteral);
        parentType = prod.parentLiteral;
      } else {
        parent = rhsNodes[prod.parentIndex];
        parentToken = parent.token;
        parentType = parent.type;
      }
      const nodeType = prod.nodeType || parentType;
      const newChildren = [];
      for (const ci of prod.withAdoptedGrandchildren) {
        const childNode = rhsNodes[ci];
        if (childNode?.children?.length) newChildren.push(...childNode.children);
      }
      return newASTNode(parentToken, nodeType, newChildren);
    }
    if (!useFullTree && prod.hasHint) {
      const nodeType = prod.nodeType || prod.lhs;
      let parentToken = null;
      if (prod.hasParentLiteral) parentToken = newToken(prod.parentLiteral, prod.parentLiteral);
      else if (prod.parentIndex >= 0 && rhsNodes[prod.parentIndex]?.token) parentToken = rhsNodes[prod.parentIndex].token;
      const hintChildren = prod.childIndices.map((ci) => rhsNodes[ci]);
      return newASTNode(parentToken, nodeType, hintChildren);
    }
    if (prod.rhsCount === 1) return rhsNodes[0];
    if (prod.rhsCount === 0) return newASTNode(null, prod.lhs, []);
    return newASTNode(null, prod.lhs, rhsNodes);
`
    : `
    if (prod.rhsCount === 1) return rhsNodes[0];
    if (prod.rhsCount === 0) return newASTNode(null, prod.lhs, []);
    return newASTNode(null, prod.lhs, rhsNodes);
`;

  return `/**
 * Generated by parsegen_code.js from parsegen JSON tables. Do not edit.
 */
import {
  newToken,
  newASTNode,
  newASTNodeTerminal,
  AST,
  TOKEN_TYPE_EOF,
  TOKEN_TYPE_ERROR,
} from "../../generators/js/runtime/index.js";

const ActionKind = { SHIFT: 0, REDUCE: 1, ACCEPT: 2 };

class Action {
  constructor(kind, target) {
    this.kind = kind;
    this.target = target;
  }
}

const ${className}_NO_AST_SENTINEL = { token: null, type: "", children: [] };

const ACTIONS = {
${actionsStr}
};

const GOTOS = {
${gotosStr}
};

const PRODUCTIONS = [
${productionsStr}
];

export class TraceHooks {
  constructor() {
    this.onToken = null;
    this.onAction = null;
    this.onStack = null;
  }
}

export class ${className} {
  constructor() {
    this.trace = null;
  }

  parse(lexer, astMode = "") {
    if (lexer == null) throw new Error("parser: nil lexer");
    const stateStack = [0];
    const nodeStack = [];
    let lookahead = lexer.scan();
    if (this.trace?.onToken && lookahead) this.trace.onToken(lookahead);
    for (;;) {
      if (lookahead == null) throw new Error("parser: lexer returned null token");
      if (lookahead.type === TOKEN_TYPE_ERROR) throw new Error(\`lexer error: \${lookahead.lexeme}\`);
      const state = stateStack[stateStack.length - 1];
      const stateActions = ACTIONS[state];
      const action = stateActions?.[lookahead.type];
      if (action == null)
        throw new Error(\`parse error: unexpected \${lookahead.type} (\${lookahead.lexeme})\`);
      if (this.trace?.onAction) this.trace.onAction(state, action, lookahead);
      if (action.kind === ActionKind.SHIFT) {
        if (astMode === "noast") nodeStack.push(${className}_NO_AST_SENTINEL);
        else nodeStack.push(newASTNodeTerminal(lookahead, lookahead.type));
        stateStack.push(action.target);
        lookahead = lexer.scan();
        if (this.trace?.onToken && lookahead) this.trace.onToken(lookahead);
        if (this.trace?.onStack) this.trace.onStack(stateStack, nodeStack);
      } else if (action.kind === ActionKind.REDUCE) {
        const prod = PRODUCTIONS[action.target];
        const rhsNodes = Array(prod.rhsCount);
        for (let i = prod.rhsCount - 1; i >= 0; i--) {
          stateStack.pop();
          rhsNodes[i] = nodeStack.pop();
        }
        if (astMode === "noast") nodeStack.push(${className}_NO_AST_SENTINEL);
        else nodeStack.push(reduce(prod, rhsNodes, astMode));
        const nextState = GOTOS[stateStack[stateStack.length - 1]]?.[prod.lhs];
        if (nextState == null) throw new Error(\`parse error: missing goto for \${prod.lhs}\`);
        stateStack.push(nextState);
        if (this.trace?.onStack) this.trace.onStack(stateStack, nodeStack);
      } else {
        if (nodeStack.length !== 1) throw new Error(\`parse error: unexpected parse stack size \${nodeStack.length}\`);
        if (this.trace?.onStack) this.trace.onStack(stateStack, nodeStack);
        if (astMode === "noast") return null;
        return new AST(nodeStack[0]);
      }
    }
  }

  attachCLITrace(traceTokens = false, traceStates = false, traceStack = false) {
    if (!traceTokens && !traceStates && !traceStack) return;
    this.trace = new TraceHooks();
    if (traceTokens) this.trace.onToken = (tok) => { if (tok) console.error(\`TOK type=\${tok.type} lexeme=\${tok.lexeme}\`); };
    if (traceStates) this.trace.onAction = (state, action, lookahead) => { console.error(\`STATE \${state} \${action.kind === ActionKind.SHIFT ? "shift" : action.kind === ActionKind.REDUCE ? "reduce" : "accept"} on \${lookahead?.type}(\${lookahead?.lexeme})\`); };
    if (traceStack) this.trace.onStack = (stateStack, nodeStack) => { console.error(\`STACK states=[\${stateStack.join(" ")}] nodes=[\${nodeStack.map((n) => n.type).join(" ")}]\`); };
  }
}

function reduce(prod, rhsNodes, astMode) {
  const useFullTree = astMode === "fullast";
${reduceBody}
}
`;
}

function parseArgs(argv) {
  const args = { output: null, className: null, prefix: "pgpg_", jsonFile: null };
  for (let i = 0; i < argv.length; i++) {
    if (argv[i] === "-o" && argv[i + 1]) {
      args.output = argv[++i];
    } else if (argv[i] === "-c" && argv[i + 1]) {
      args.className = argv[++i];
    } else if (argv[i] === "--prefix" && argv[i + 1]) {
      args.prefix = argv[++i];
    } else if (!argv[i].startsWith("-")) {
      args.jsonFile = argv[i];
    }
  }
  return args;
}

function main() {
  const argv = process.argv.slice(2);
  const args = parseArgs(argv);
  if (!args.output || !args.className || !args.jsonFile) {
    console.error("Usage: node parsegen_code.js -o <output.js> -c <ClassName> [--prefix pgpg_] <parse.json>");
    process.exit(1);
  }
  const raw = loadTables(args.jsonFile);
  const hintMode = raw.hint_mode ?? "";
  const className = args.prefix ? args.prefix + args.className : args.className;
  const actions = buildActions(raw);
  const gotos = buildGotos(raw);
  const productions = buildProductions(raw);
  const outDir = path.dirname(args.output);
  fs.mkdirSync(outDir, { recursive: true });
  fs.writeFileSync(
    args.output,
    renderParser({ className, actions, gotos, productions, hintMode }),
    "utf-8",
  );
}

main();

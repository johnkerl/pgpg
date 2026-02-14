import { ASTNode } from "./ast-node.js";

export function newASTNode(token, nodeType, children) {
  return new ASTNode(token, nodeType, children ?? []);
}

export function newASTNodeTerminal(token, nodeType) {
  return new ASTNode(token, nodeType, []);
}

export class AST {
  constructor(rootNode) {
    this.rootNode = rootNode;
  }

  printTree(indent = 0) {
    const pad = "    ".repeat(indent);
    const node = this.rootNode;
    if (node.token) {
      const lexeme = node.token.lexeme.replace(/\n/g, "\\n");
      console.log(`${pad}${JSON.stringify(lexeme)} [tt:${node.token.type}] [nt:${node.type}]`);
    } else {
      console.log(`${pad}[nt:${node.type}]`);
    }
    for (const child of node.children) {
      new AST(child).printTree(indent + 1);
    }
  }
}

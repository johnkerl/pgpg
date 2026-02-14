/**
 * A single node in the abstract syntax tree.
 */
export class ASTNode {
  constructor(token, type, children = []) {
    this.token = token;
    this.type = type;
    this.children = children ?? [];
  }
}

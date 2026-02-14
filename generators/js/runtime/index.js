/**
 * PGPG JavaScript runtime: Token, AST, and lexer contract for generated parsers/lexers.
 */
export {
  Location,
  Token,
  TOKEN_TYPE_ERROR,
  TOKEN_TYPE_EOF,
  newToken,
  newEOFToken,
  newErrorToken,
} from "./token.js";
export { AST, ASTNode, newASTNode, newASTNodeTerminal } from "./ast.js";

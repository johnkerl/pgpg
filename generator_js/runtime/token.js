/**
 * Token and location types for lexer/parser.
 */

export class Location {
  constructor(line = 1, column = 1, byteOffset = 0) {
    this.line = line;
    this.column = column;
    this.byteOffset = byteOffset;
  }
}

export const TOKEN_TYPE_ERROR = "!error";
export const TOKEN_TYPE_EOF = "EOF";

export class Token {
  constructor(type, lexeme, location = null) {
    this.type = type;
    this.lexeme = lexeme;
    this.location = location;
  }
}

export function newToken(lexeme, tokenType, location = null) {
  return new Token(tokenType, lexeme, location);
}

export function newEOFToken(location = null) {
  return new Token(TOKEN_TYPE_EOF, "", location);
}

export function newErrorToken(message, location = null) {
  return new Token(TOKEN_TYPE_ERROR, message, location);
}

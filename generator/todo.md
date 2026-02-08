* DFA minimization
* Maybe: auto‑include operator literals from parser rules:
  * Scan all grammar rules
  * Collect any literal terminals (e.g., "+", "-", "(", ")", "==", etc.).
  * Add those literals into the lexer rule set automatically, even if no explicit lexer rule exists


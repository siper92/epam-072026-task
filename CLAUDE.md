# CLAUDE

info [README.md](README.md)

# Code
 - don't use to long lines, prefer exceptions or other structures to be on multiple lines
   - human centric formating is preferred - break lines by meaning
 - always add default values to .env.default

# do
- prefer usage of tools ro examine and analyze code and dependencies
- make usage of tools required for sub-agents
   - sub agent's will write back reports to main agent main agent
- use SOLID principles and best practices

# Test
 - table-driven where sensible, using t.Run, 
 - plain testing package, and the error-code helper errs.HasCode from ticTacSolved/task/pkg/errs (already imported). 
 - Known error codes used by the package: CodeInvalidInput, CodeInvalidTransition, CodeGameFinished, CodeOutOfBounds, CodeCellOccupied.

# Documents and MD files
- use only askii symbols and `-` never any other characters or emojis (for emojis exceptions are allowed when asked for)
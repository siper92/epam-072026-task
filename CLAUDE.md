# CLAUDE

info [README.md](README.md)

# Code
 - don't use to long lines, prefer exceptions or other structures to be on multiple lines
   - human centric formating is preferred - break lines by meaning

# Test
 - table-driven where sensible, using t.Run, 
 - plain testing package, and the error-code helper errs.HasCode from epam/task/pkg/errs (already imported). 
 - Known error codes used by the package: CodeInvalidInput, CodeInvalidTransition, CodeGameFinished, CodeOutOfBounds, CodeCellOccupied.

# Documents and MD files
- use only askii symbols and `-` never any other characters or emojis (for emojis exceptions are allowed when asked for)
const vscode = require("vscode");
const cp = require("child_process");
const path = require("path");
const fs = require("fs");

/**
 * Locate the framework CLI executable or entry-point for the current
 * workspace.  Priority:
 *   1. User setting   (gonx.formatterPath)
 *   2. Workspace root /scripts/framework/main.go  (=> "go run …")
 *   3. Binary "framework" on $PATH
 */
function resolveFormatter() {
  const cfg = vscode.workspace.getConfiguration("gonx");
  const userPath = cfg.get("formatterPath");
  if (userPath) {
    return { cmd: userPath, args: [] };
  }

  const ws = vscode.workspace.workspaceFolders?.[0];
  if (ws) {
    const candidate = path.join(ws.uri.fsPath, "scripts", "framework", "main.go");
    if (fs.existsSync(candidate)) {
      return { cmd: "go", args: ["run", candidate] };
    }
  }

  return { cmd: "framework", args: [] };
}

/**
 * Build the argument list:  [ …baseArgs, "fmt", filePath ]
 */
function buildArgs(baseArgs, filePath) {
  return [...baseArgs, "fmt", filePath];
}

class GonxFormatProvider {
  provideDocumentFormattingEdits(document) {
    return new Promise((resolve, reject) => {
      const filePath = document.uri.fsPath;
      const { cmd, args } = resolveFormatter();
      const allArgs = buildArgs(args, filePath);

      cp.execFile(cmd, allArgs, { cwd: path.dirname(filePath) }, (err, stdout, stderr) => {
        if (err) {
          const msg = stderr?.trim() || err.message;
          vscode.window.showErrorMessage(`Gonx formatter error: ${msg}`);
          return reject(new Error(msg));
        }

        // After formatting, read the file back and compute a full-document replace
        fs.readFile(filePath, "utf8", (readErr, newText) => {
          if (readErr) {
            return reject(readErr);
          }
          const fullRange = new vscode.Range(
            document.positionAt(0),
            document.positionAt(document.getText().length)
          );
          resolve([new vscode.TextEdit(fullRange, newText)]);
        });
      });
    });
  }
}

function activate(context) {
  const provider = vscode.languages.registerDocumentFormattingEditProvider(
    "gonx",
    new GonxFormatProvider()
  );
  context.subscriptions.push(provider);
}

function deactivate() {}

module.exports = { activate, deactivate };

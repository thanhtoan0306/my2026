import * as vscode from "vscode";
import { beautifyJson } from "./jsonBeautify";

function getOptions(): { sortKeys: boolean; indent: number } {
  const cfg = vscode.workspace.getConfiguration("terminalJsonBeautify");
  return {
    sortKeys: cfg.get<boolean>("sortKeys", true),
    indent: cfg.get<number>("indent", 2),
  };
}

async function readSelectionText(): Promise<string | undefined> {
  const editor = vscode.window.activeTextEditor;
  if (editor && !editor.selection.isEmpty) {
    return editor.document.getText(editor.selection);
  }
  return vscode.env.clipboard.readText();
}

async function readTerminalSelection(): Promise<string | undefined> {
  const terminal = vscode.window.activeTerminal;
  if (!terminal) {
    vscode.window.showWarningMessage("No active terminal.");
    return undefined;
  }

  await vscode.commands.executeCommand("workbench.action.terminal.copySelection");
  const text = await vscode.env.clipboard.readText();
  if (!text.trim()) {
    vscode.window.showWarningMessage(
      "Select JSON in the terminal first, then run the command again."
    );
    return undefined;
  }
  return text;
}

async function showBeautified(text: string): Promise<void> {
  const cfg = vscode.workspace.getConfiguration("terminalJsonBeautify");
  const openInEditor = cfg.get<boolean>("openInEditor", true);

  if (openInEditor) {
    const doc = await vscode.workspace.openTextDocument({
      content: text,
      language: "json",
    });
    await vscode.window.showTextDocument(doc, { preview: false });
  }

  await vscode.env.clipboard.writeText(text);
  vscode.window.showInformationMessage(
    "JSON beautified and copied to clipboard."
  );
}

async function runBeautify(source: string): Promise<void> {
  const result = beautifyJson(source, getOptions());
  if (!result.ok) {
    vscode.window.showErrorMessage(result.error);
    return;
  }
  await showBeautified(result.text);
}

export function activate(context: vscode.ExtensionContext): void {
  context.subscriptions.push(
    vscode.commands.registerCommand(
      "terminalJsonBeautify.selection",
      async () => {
        const text = await readSelectionText();
        if (text === undefined) {
          return;
        }
        await runBeautify(text);
      }
    ),
    vscode.commands.registerCommand(
      "terminalJsonBeautify.terminal",
      async () => {
        const text = await readTerminalSelection();
        if (text === undefined) {
          return;
        }
        await runBeautify(text);
      }
    ),
    vscode.commands.registerCommand(
      "terminalJsonBeautify.clipboard",
      async () => {
        const text = await vscode.env.clipboard.readText();
        await runBeautify(text);
      }
    ),
    vscode.commands.registerCommand(
      "terminalJsonBeautify.editor",
      async () => {
        const editor = vscode.window.activeTextEditor;
        if (!editor) {
          vscode.window.showWarningMessage("Open a file or terminal output first.");
          return;
        }
        const range = editor.selection.isEmpty
          ? new vscode.Range(
              editor.document.positionAt(0),
              editor.document.positionAt(editor.document.getText().length)
            )
          : editor.selection;
        const text = editor.document.getText(range);
        const result = beautifyJson(text, getOptions());
        if (!result.ok) {
          vscode.window.showErrorMessage(result.error);
          return;
        }
        await editor.edit((eb) => eb.replace(range, result.text.trimEnd()));
        vscode.window.showInformationMessage("JSON beautified in editor.");
      }
    )
  );
}

export function deactivate(): void {}

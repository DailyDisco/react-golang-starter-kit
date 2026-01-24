import { useCallback, useState } from "react";

import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { getModifierKey, useKeyboardShortcuts } from "@/hooks/useKeyboardShortcuts";
import { Keyboard } from "lucide-react";
import { useTranslation } from "react-i18next";

interface ShortcutGroup {
  title: string;
  shortcuts: {
    keys: string;
    description: string;
  }[];
}

const shortcutGroups: ShortcutGroup[] = [
  {
    title: "General",
    shortcuts: [
      { keys: `${getModifierKey()}+K`, description: "Open command palette" },
      { keys: "Escape", description: "Close dialogs/modals" },
      { keys: `${getModifierKey()}+,`, description: "Open settings" },
    ],
  },
  {
    title: "Navigation",
    shortcuts: [
      { keys: `${getModifierKey()}+/`, description: "Show keyboard shortcuts" },
      { keys: "G then D", description: "Go to dashboard" },
      { keys: "G then S", description: "Go to settings" },
      { keys: "G then B", description: "Go to billing" },
    ],
  },
  {
    title: "Actions",
    shortcuts: [
      { keys: `${getModifierKey()}+S`, description: "Save (when editing)" },
      { keys: `${getModifierKey()}+Enter`, description: "Submit form" },
      { keys: `${getModifierKey()}+Z`, description: "Undo" },
      { keys: `${getModifierKey()}+Shift+Z`, description: "Redo" },
    ],
  },
  {
    title: "Table Navigation",
    shortcuts: [
      { keys: "Arrow Keys", description: "Navigate cells" },
      { keys: "Enter", description: "Edit cell / Confirm" },
      { keys: "Space", description: "Toggle selection" },
      { keys: `${getModifierKey()}+A`, description: "Select all" },
    ],
  },
];

export function KeyboardShortcutsHelp() {
  const { t } = useTranslation("common");
  const [isOpen, setIsOpen] = useState(false);

  const toggle = useCallback(() => setIsOpen((prev) => !prev), []);

  // Register multiple shortcuts to open this dialog
  useKeyboardShortcuts({
    shortcuts: [
      {
        key: "/",
        meta: true,
        handler: toggle,
        description: "Show keyboard shortcuts",
      },
      {
        key: "?",
        shift: true,
        handler: toggle,
        description: "Show keyboard shortcuts",
      },
    ],
  });

  return (
    <Dialog
      open={isOpen}
      onOpenChange={setIsOpen}
    >
      <DialogTrigger asChild>
        <Button
          variant="ghost"
          size="icon"
          className="h-9 w-9"
          title="Keyboard shortcuts"
        >
          <Keyboard className="h-4 w-4" />
        </Button>
      </DialogTrigger>
      <DialogContent className="max-w-2xl">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Keyboard className="h-5 w-5" />
            {t("shortcuts.title", "Keyboard Shortcuts")}
          </DialogTitle>
          <DialogDescription>
            {t("shortcuts.description", "Use these shortcuts to navigate and take actions quickly.")}
          </DialogDescription>
        </DialogHeader>

        <div className="mt-4 grid gap-6 md:grid-cols-2">
          {shortcutGroups.map((group) => (
            <div key={group.title}>
              <h3 className="text-muted-foreground mb-3 text-sm font-medium">{group.title}</h3>
              <div className="space-y-2">
                {group.shortcuts.map((shortcut) => (
                  <div
                    key={shortcut.description}
                    className="flex items-center justify-between"
                  >
                    <span className="text-sm">{shortcut.description}</span>
                    <kbd className="bg-muted rounded px-2 py-1 font-mono text-xs">{shortcut.keys}</kbd>
                  </div>
                ))}
              </div>
            </div>
          ))}
        </div>

        <div className="bg-muted/50 mt-6 rounded-lg p-3 text-center">
          <p className="text-muted-foreground text-sm">
            {t("shortcuts.tip", "Press {{key}} anytime to open the command palette", {
              key: `${getModifierKey()}+K`,
            })}
          </p>
        </div>
      </DialogContent>
    </Dialog>
  );
}

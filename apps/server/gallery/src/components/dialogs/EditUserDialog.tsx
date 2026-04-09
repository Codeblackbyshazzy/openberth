import { useState, useEffect } from "react";
import { Button } from "../ui/button";
import { Input } from "../ui/input";
import { Label } from "../ui/label";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "../ui/dialog";

export interface UserInfo {
  name: string;
  displayName: string;
  role: string;
  maxDeployments: number;
  createdAt: string;
}

interface EditUserDialogProps {
  target: UserInfo | null;
  onClose: () => void;
  onSave: (name: string, data: { displayName?: string; password?: string; maxDeployments?: number }) => Promise<void>;
}

export function EditUserDialog({ target, onClose, onSave }: EditUserDialogProps) {
  const [displayName, setDisplayName] = useState("");
  const [password, setPassword] = useState("");
  const [maxDeploys, setMaxDeploys] = useState("");
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    if (target) {
      setDisplayName(target.displayName || "");
      setPassword("");
      setMaxDeploys(String(target.maxDeployments));
    }
  }, [target]);

  const handleSave = async () => {
    if (!target) return;
    setSaving(true);
    const body: { displayName?: string; password?: string; maxDeployments?: number } = {};
    if (displayName !== (target.displayName || "")) body.displayName = displayName;
    if (password) body.password = password;
    const maxVal = parseInt(maxDeploys, 10);
    if (!isNaN(maxVal) && maxVal !== target.maxDeployments) body.maxDeployments = maxVal;
    await onSave(target.name, body);
    setSaving(false);
    onClose();
  };

  return (
    <Dialog open={!!target} onOpenChange={(open) => !open && onClose()}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Edit user</DialogTitle>
          <DialogDescription>
            Update settings for <strong>{target?.name}</strong>.
          </DialogDescription>
        </DialogHeader>
        <div className="grid gap-4 py-2">
          <div className="grid gap-2">
            <Label htmlFor="eu-display">Display Name</Label>
            <Input id="eu-display" value={displayName} onChange={(e) => setDisplayName(e.target.value)} placeholder="Optional display name" />
          </div>
          <div className="grid gap-2">
            <Label htmlFor="eu-max">Max Deployments</Label>
            <Input id="eu-max" type="number" min="0" value={maxDeploys} onChange={(e) => setMaxDeploys(e.target.value)} />
          </div>
          <div className="grid gap-2">
            <Label htmlFor="eu-pass">New Password</Label>
            <Input id="eu-pass" type="password" value={password} onChange={(e) => setPassword(e.target.value)} placeholder="Leave empty to keep current" />
            {password && password.length < 8 && <p className="text-xs text-destructive">Minimum 8 characters</p>}
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={onClose} disabled={saving}>Cancel</Button>
          <Button onClick={handleSave} disabled={saving || (password.length > 0 && password.length < 8)}>
            {saving ? "Saving..." : "Save"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

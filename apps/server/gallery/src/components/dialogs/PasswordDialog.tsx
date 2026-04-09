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

interface PasswordDialogProps {
  open: boolean;
  hasPassword: boolean;
  onClose: () => void;
  onSave: (currentPassword: string, newPassword: string) => Promise<string | null>;
}

import { useState } from "react";

export function PasswordDialog({ open, hasPassword, onClose, onSave }: PasswordDialogProps) {
  const [currentPassword, setCurrentPassword] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");

  const handleClose = () => {
    setCurrentPassword("");
    setNewPassword("");
    setConfirmPassword("");
    setError("");
    onClose();
  };

  const handleSave = async () => {
    if (newPassword !== confirmPassword) {
      setError("Passwords do not match");
      return;
    }
    if (newPassword.length < 8) {
      setError("Password must be at least 8 characters");
      return;
    }
    setSaving(true);
    setError("");
    const err = await onSave(currentPassword, newPassword);
    setSaving(false);
    if (err) {
      setError(err);
    } else {
      handleClose();
    }
  };

  return (
    <Dialog open={open} onOpenChange={(o) => !o && handleClose()}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>{hasPassword ? "Change password" : "Set password"}</DialogTitle>
          <DialogDescription>
            {hasPassword
              ? "Enter your current password and choose a new one."
              : "Set a password to log in via the web dashboard."}
          </DialogDescription>
        </DialogHeader>
        <div className="grid gap-4 py-2">
          {hasPassword && (
            <div className="grid gap-2">
              <Label htmlFor="pw-current">Current password</Label>
              <Input id="pw-current" type="password" value={currentPassword} onChange={(e) => setCurrentPassword(e.target.value)} />
            </div>
          )}
          <div className="grid gap-2">
            <Label htmlFor="pw-new">New password</Label>
            <Input id="pw-new" type="password" value={newPassword} onChange={(e) => setNewPassword(e.target.value)} />
            {newPassword && newPassword.length < 8 && <p className="text-xs text-destructive">Minimum 8 characters</p>}
          </div>
          <div className="grid gap-2">
            <Label htmlFor="pw-confirm">Confirm new password</Label>
            <Input id="pw-confirm" type="password" value={confirmPassword} onChange={(e) => setConfirmPassword(e.target.value)} />
            {confirmPassword && newPassword !== confirmPassword && <p className="text-xs text-destructive">Passwords do not match</p>}
          </div>
          {error && <p className="text-sm text-destructive">{error}</p>}
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={handleClose} disabled={saving}>Cancel</Button>
          <Button onClick={handleSave} disabled={saving || newPassword.length < 8 || newPassword !== confirmPassword}>
            {saving ? "Saving..." : "Save"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

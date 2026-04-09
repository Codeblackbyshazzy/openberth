import { useState } from "react";
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

interface CreateUserDialogProps {
  open: boolean;
  onClose: () => void;
  onCreate: (name: string, password: string, maxDeploys: string) => Promise<void>;
}

export function CreateUserDialog({ open, onClose, onCreate }: CreateUserDialogProps) {
  const [name, setName] = useState("");
  const [password, setPassword] = useState("");
  const [maxDeploys, setMaxDeploys] = useState("");
  const [creating, setCreating] = useState(false);

  const handleClose = () => {
    setName("");
    setPassword("");
    setMaxDeploys("");
    onClose();
  };

  const handleCreate = async () => {
    setCreating(true);
    await onCreate(name, password, maxDeploys);
    setCreating(false);
    handleClose();
  };

  return (
    <Dialog open={open} onOpenChange={(o) => !o && handleClose()}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Create user</DialogTitle>
          <DialogDescription>Add a new user to OpenBerth.</DialogDescription>
        </DialogHeader>
        <div className="grid gap-4 py-2">
          <div className="grid gap-2">
            <Label htmlFor="new-user-name">Username</Label>
            <Input id="new-user-name" value={name} onChange={(e) => setName(e.target.value)} />
          </div>
          <div className="grid gap-2">
            <Label htmlFor="new-user-pass">Password (optional)</Label>
            <Input id="new-user-pass" type="password" value={password} onChange={(e) => setPassword(e.target.value)} placeholder="Leave empty for API key only" />
          </div>
          <div className="grid gap-2">
            <Label htmlFor="new-user-max">Max Deployments</Label>
            <Input id="new-user-max" type="number" min="0" value={maxDeploys} onChange={(e) => setMaxDeploys(e.target.value)} placeholder="Default from server config" />
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={handleClose} disabled={creating}>Cancel</Button>
          <Button onClick={handleCreate} disabled={creating || !name}>{creating ? "Creating..." : "Create"}</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

import { useState, useEffect } from "react";
import { KeyRound } from "lucide-react";
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
import type { GalleryItem } from "../../types";

interface EditDialogProps {
  target: GalleryItem | null;
  onClose: () => void;
  onSave: (data: EditData) => Promise<void>;
}

export interface EditData {
  title: string;
  description: string;
  ttl: string;
  networkQuota: string;
  accessMode: string;
  username: string;
  password: string;
  apiKey: string;
  accessUsers: string;
}

export function EditDialog({ target, onClose, onSave }: EditDialogProps) {
  const [title, setTitle] = useState("");
  const [desc, setDesc] = useState("");
  const [ttl, setTTL] = useState("0");
  const [accessMode, setAccessMode] = useState("public");
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [apiKey, setApiKey] = useState("");
  const [networkQuota, setNetworkQuota] = useState("");
  const [accessUsers, setAccessUsers] = useState("");
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    if (target) {
      setTitle(target.title);
      setDesc(target.description);
      const ttlMap: Record<number, string> = { 0: "0", 24: "24h", 72: "72h", 168: "7d", 720: "30d" };
      setTTL(ttlMap[target.ttlHours] ?? "0");
      setAccessMode(target.accessMode || "public");
      setUsername(target.accessUser || "");
      setPassword("");
      setApiKey("");
      setNetworkQuota(target.networkQuota || "");
      setAccessUsers(target.accessUsers || "");
    }
  }, [target]);

  const handleSave = async () => {
    setSaving(true);
    await onSave({ title, description: desc, ttl, networkQuota, accessMode, username, password, apiKey, accessUsers });
    setSaving(false);
  };

  const selectClass = "flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2";

  return (
    <Dialog open={!!target} onOpenChange={(open) => !open && onClose()}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Edit deployment</DialogTitle>
          <DialogDescription>
            Update metadata for <strong>{target?.title || target?.name}</strong>.
          </DialogDescription>
        </DialogHeader>
        <div className="grid gap-4 py-2">
          <div className="grid gap-2">
            <Label htmlFor="edit-title">Title</Label>
            <Input id="edit-title" value={title} onChange={(e) => setTitle(e.target.value)} />
          </div>
          <div className="grid gap-2">
            <Label htmlFor="edit-desc">Description</Label>
            <Input id="edit-desc" value={desc} onChange={(e) => setDesc(e.target.value)} />
          </div>
          <div className="grid gap-2">
            <Label htmlFor="edit-ttl">TTL</Label>
            <select id="edit-ttl" value={ttl} onChange={(e) => setTTL(e.target.value)} className={selectClass}>
              <option value="0">Never expires</option>
              <option value="24h">24 hours</option>
              <option value="72h">3 days</option>
              <option value="7d">7 days</option>
              <option value="30d">30 days</option>
            </select>
          </div>
          <div className="grid gap-2">
            <Label htmlFor="edit-quota">Network quota</Label>
            <Input id="edit-quota" value={networkQuota} onChange={(e) => setNetworkQuota(e.target.value)} placeholder="e.g. 1g, 5g (empty = no quota)" />
            <p className="text-xs text-muted-foreground">Leave empty to use server default or disable</p>
          </div>
          <div className="grid gap-2">
            <Label htmlFor="edit-access">Protection</Label>
            <select id="edit-access" value={accessMode} onChange={(e) => setAccessMode(e.target.value)} className={selectClass}>
              <option value="public">Public</option>
              <option value="basic_auth">Basic Auth</option>
              <option value="api_key">API Key</option>
              <option value="user">User (Login Required)</option>
            </select>
          </div>
          {accessMode === "basic_auth" && (
            <div className="grid gap-2">
              <Label htmlFor="edit-user">Username</Label>
              <Input id="edit-user" value={username} onChange={(e) => setUsername(e.target.value)} />
              <Label htmlFor="edit-pass">Password</Label>
              <Input id="edit-pass" type="password" value={password} onChange={(e) => setPassword(e.target.value)} placeholder={target?.accessMode === "basic_auth" ? "Leave empty to keep current" : ""} />
              {target?.accessMode === "basic_auth" && !password && (
                <p className="text-xs text-muted-foreground">Currently configured. Enter a new password to update.</p>
              )}
            </div>
          )}
          {accessMode === "api_key" && (
            <div className="grid gap-2">
              <Label htmlFor="edit-ak">API Key</Label>
              <div className="flex items-center gap-2">
                <KeyRound className="h-4 w-4 text-muted-foreground shrink-0" />
                <Input id="edit-ak" value={apiKey} onChange={(e) => setApiKey(e.target.value)} placeholder={target?.accessMode === "api_key" ? "Leave empty to keep current" : "Leave empty to auto-generate"} />
              </div>
              {target?.accessMode === "api_key" && !apiKey && (
                <p className="text-xs text-muted-foreground">Currently configured. Enter a new key to update.</p>
              )}
            </div>
          )}
          {accessMode === "user" && (
            <div className="grid gap-2">
              <Label htmlFor="edit-access-users">Allowed Users</Label>
              <Input id="edit-access-users" value={accessUsers} onChange={(e) => setAccessUsers(e.target.value)} placeholder="All authenticated users" />
              <p className="text-xs text-muted-foreground">Comma-separated usernames. Leave empty to allow all authenticated users. The deployment owner and admins always have access.</p>
            </div>
          )}
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={onClose} disabled={saving}>Cancel</Button>
          <Button onClick={handleSave} disabled={saving}>{saving ? "Saving..." : "Save"}</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

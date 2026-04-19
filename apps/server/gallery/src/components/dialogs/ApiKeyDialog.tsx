import { Button } from "../ui/button";
import { Label } from "../ui/label";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "../ui/dialog";

interface ApiKeyDialogProps {
  data: { name: string; key: string } | null;
  variant: "deployment" | "user" | "user-rotated";
  onClose: () => void;
}

export function ApiKeyDialog({ data, variant, onClose }: ApiKeyDialogProps) {
  let title: string;
  switch (variant) {
    case "deployment": title = "API key protection enabled"; break;
    case "user-rotated": title = "API key rotated"; break;
    default: title = "User created";
  }
  const nameLabel = variant === "deployment" ? "Deployment" : "Username";

  return (
    <Dialog open={!!data} onOpenChange={(open) => !open && onClose()}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>{title}</DialogTitle>
          <DialogDescription>Save this API key now — it won't be shown again.</DialogDescription>
        </DialogHeader>
        <div className="grid gap-3 py-2">
          <div className="grid gap-1">
            <Label className="text-xs text-muted-foreground">{nameLabel}</Label>
            <p className="font-medium">{data?.name}</p>
          </div>
          <div className="grid gap-1">
            <Label className="text-xs text-muted-foreground">API Key</Label>
            <div className="flex items-center gap-2">
              <code className="flex-1 rounded bg-muted px-3 py-2 text-sm font-mono break-all">{data?.key}</code>
              <Button variant="outline" size="sm" onClick={() => data?.key && navigator.clipboard.writeText(data.key)}>Copy</Button>
            </div>
          </div>
          {variant === "deployment" && (
            <div className="rounded-md border bg-muted/50 p-3 text-xs text-muted-foreground space-y-1">
              <p>Use this key to access the deployment via header:</p>
              <p><code className="bg-background px-1 rounded">X-Api-Key: {data?.key}</code></p>
            </div>
          )}
        </div>
        <DialogFooter>
          <Button onClick={onClose}>Done</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

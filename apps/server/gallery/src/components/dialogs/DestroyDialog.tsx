import { Button } from "../ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "../ui/dialog";
import type { GalleryItem } from "../../types";

interface DestroyDialogProps {
  target: GalleryItem | null;
  destroying: boolean;
  onClose: () => void;
  onConfirm: () => void;
}

export function DestroyDialog({ target, destroying, onClose, onConfirm }: DestroyDialogProps) {
  return (
    <Dialog open={!!target} onOpenChange={(open) => !open && onClose()}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Destroy deployment</DialogTitle>
          <DialogDescription>
            Are you sure you want to destroy <strong>{target?.title || target?.name}</strong>? This action cannot be undone.
          </DialogDescription>
        </DialogHeader>
        <DialogFooter>
          <Button variant="outline" onClick={onClose} disabled={destroying}>
            Cancel
          </Button>
          <Button variant="destructive" onClick={onConfirm} disabled={destroying}>
            {destroying ? "Destroying..." : "Destroy"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

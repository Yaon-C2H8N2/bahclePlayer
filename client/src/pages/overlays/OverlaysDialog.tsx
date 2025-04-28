import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogHeader,
    DialogTitle
} from "@/components/ui/dialog.tsx";
import { Overlays } from "@/pages/overlays/Overlays.tsx";

interface IOverlaysDialogProps {
    open: boolean
    onClose: () => void
}

export const OverlaysDialog = (props: IOverlaysDialogProps) => {
    return (
        <Dialog open={props.open} onOpenChange={(open) => !open && props.onClose()}>
            <DialogContent className="max-w-4xl">
                <DialogHeader>
                    <DialogTitle>Overlays</DialogTitle>
                    <DialogDescription>Manage your stream overlays.</DialogDescription>
                </DialogHeader>
                <Overlays onClose={props.onClose} />
            </DialogContent>
        </Dialog>
    )
}

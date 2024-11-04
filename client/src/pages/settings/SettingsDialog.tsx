import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogHeader,
    DialogTitle
} from "@/components/ui/dialog.tsx";
import {Settings} from "@/pages/settings/Settings.tsx";

interface ISettingsDialogProps {
    open: boolean
    onClose: () => void
}

export const SettingsDialog = (props: ISettingsDialogProps) => {

    return (
        <Dialog open={props.open} onOpenChange={(open) => !open && props.onClose()}>
            <DialogContent>
                <DialogHeader>
                    <DialogTitle>Settings</DialogTitle>
                    <DialogDescription>Set up which reward to use to add track to your playlist and queue.</DialogDescription>
                </DialogHeader>
                <Settings onClose={props.onClose}/>
            </DialogContent>
        </Dialog>
    )
}
import {Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle} from "@/components/ui/dialog.tsx";
import {Input} from "@/components/ui/input.tsx";
import {Select, SelectContent, SelectItem, SelectTrigger, SelectValue} from "@/components/ui/select.tsx";
import {useState} from "react";
import {Button} from "@/components/ui/button.tsx";
import {fetchApi} from "@/lib/network.ts";
import {useToast} from "@/hooks/use-toast.ts";

interface IManualAddDialogProps {
    open: boolean
    onClose: () => void
}

export const ManualAddDialog = (props: IManualAddDialogProps) => {
    const [type, setType] = useState<"QUEUE" | "PLAYLIST">("QUEUE")
    const [videoUrl, setVideoUrl] = useState<string>("")
    const [loading, setLoading] = useState<boolean>(false)
    const toast = useToast()

    const handleSave = async () => {
        setLoading(true)

        const res = await fetchApi("/api/addVideo", {
            method: "POST",
            body: JSON.stringify({
                url: videoUrl,
                type: type
            })
        })

        if (res.ok) {
            props.onClose()
        } else {
            const data = await res.json()
            toast.toast({
                title: "Error",
                description: data.error,
                variant: "destructive",
            })
        }

        setLoading(false)
    }

    return (
        <Dialog open={props.open} onOpenChange={(open) => !open && props.onClose()}>
            <DialogContent>
                <DialogHeader>
                    <DialogTitle>Add a track</DialogTitle>
                    <DialogDescription>Add a track to the queue or playlist.</DialogDescription>
                </DialogHeader>
                <div>
                    <div>Video URL :</div>
                    <Input onChange={(event) => {
                        setVideoUrl(event.target.value)
                    }}/>
                </div>
                <div>
                    <div>Add to :</div>
                    <Select value={type} onValueChange={(value) => setType(value as "QUEUE" | "PLAYLIST")}>
                        <SelectTrigger>
                            <SelectValue placeholder="Mode"/>
                        </SelectTrigger>
                        <SelectContent>
                            <SelectItem value="QUEUE">Queue</SelectItem>
                            <SelectItem value="PLAYLIST">Playlist</SelectItem>
                        </SelectContent>
                    </Select>
                </div>
                <div className={"flex flex-row justify-end gap-2 w-full"}>
                    <Button variant={"ghost"} onClick={() => props.onClose()}>Cancel</Button>
                    <Button disabled={loading} onClick={() => handleSave()}>Save</Button>
                </div>
            </DialogContent>
        </Dialog>
    )
}
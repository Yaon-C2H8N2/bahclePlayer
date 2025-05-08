import {Card} from "@/components/ui/card.tsx";
import {Play, Trash} from "lucide-react";

interface IVideoCardProps {
    video: Video
    isPlaying?: boolean
    onRemove?: (video: Video) => void
}

export const VideoCard = (props: IVideoCardProps) => {
    return (
        <Card className={"max-w-[100%] min-w-[100%] min-h-[7rem] max-h-[7rem] flex justify-between items-center px-3 group"}>
            <div className={"flex flex-row gap-3"}>
                <div className={"max-h-full min-w-32 max-w-32"}>
                    <img width={"100%"} height={"auto"} src={props.video.thumbnail_url}></img>
                </div>
                <div className={"flex flex-col"}>
                    <div className={"line-clamp-2"}>{props.video.title}</div>
                    <div>{props.video.duration}</div>
                    <div>Added by {props.video.added_by}</div>
                </div>
            </div>
            {props.isPlaying && <Play className={"m-3 min-w-5 max-w-5"}/>}
            <Trash onClick={()=>{props.onRemove && props.onRemove(props.video)}} color={"#FF0000"} className={"m-3 min-w-5 max-w-5 hidden hover:cursor-pointer group-hover:block"}/>
        </Card>
    )
}
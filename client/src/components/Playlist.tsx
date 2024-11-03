import {Card} from "@/components/ui/card.tsx";
import {Play} from "lucide-react";

interface IPlaylistProps {
    playlist: any[]
    queue: any[]
    currentlyPlaying?: any
}

interface IVideoCardProps {
    video: any
    isPlaying?: boolean
}

const VideoCard = (props: IVideoCardProps) => {
    return (
        <Card className={"max-w-[100%] min-w-[100%] min-h-[7rem] max-h-[7rem] flex justify-between items-center px-3"}>
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
        </Card>
    )
}

export const Playlist = (props: IPlaylistProps) => {
    return (
        <div className={"flex flex-col flex-1 w-full justify-center items-center gap-3"}>
            <div>Queue{props.queue.length > 0 && " - ("+props.queue.length+" tracks)"}</div>
            {props.queue.length === 0 && <div>Queue is empty</div>}
            {props.queue.map((video: any) => {
                return (
                    <VideoCard key={video.video_id} video={video} isPlaying={props?.currentlyPlaying === video}/>
                )
            })}
            <div>Playlist{props.playlist.length > 0 && " - (" + props.playlist.length + " tracks)"}</div>
            {props.playlist.length === 0 && <div>Playlist is empty</div>}
            {props.playlist.map((video: any) => {
                return (
                    <VideoCard key={video.video_id} video={video} isPlaying={props?.currentlyPlaying === video}/>
                )
            })}
        </div>
    )
}
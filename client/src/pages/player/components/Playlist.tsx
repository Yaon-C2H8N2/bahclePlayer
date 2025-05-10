import {Accordion, AccordionContent, AccordionItem, AccordionTrigger} from "@/components/ui/accordion.tsx";
import {VideoCard} from "@/pages/player/components/VideoCard.tsx";

interface IPlaylistProps {
    playlist: Video[]
    queue: Video[]
    currentlyPlaying?: Video
    onRemoveVideo?: (video: Video) => void
}

export const Playlist = (props: IPlaylistProps) => {
    return (
        <div className={"flex flex-col flex-1 min-w-96 max-w-[100%] mr-5 justify-center items-center gap-3"}>
            <Accordion type={"single"} collapsible={true} className={"w-full"}>
                <AccordionItem value={"queue"}>
                    <AccordionTrigger>Queue{props.queue.length > 0 && " - ("+props.queue.length+" tracks)"}</AccordionTrigger>
                    <AccordionContent>
                        {props.queue.length === 0 && <div>Queue is empty</div>}
                        {props.queue.map((video) => {
                            return (
                                <VideoCard key={video.video_id} video={video} isPlaying={props?.currentlyPlaying === video} onRemove={props.onRemoveVideo}/>
                            )
                        })}
                    </AccordionContent>
                </AccordionItem>
            </Accordion>
            <Accordion type={"single"} collapsible={true} className={"w-full"}>
                <AccordionItem value={"playlist"}>
                    <AccordionTrigger>Playlist{props.playlist.length > 0 && " - (" + props.playlist.length + " tracks)"}</AccordionTrigger>
                    <AccordionContent>
                        {props.playlist.length === 0 && <div>Playlist is empty</div>}
                        {props.playlist.map((video) => {
                            return (
                                <VideoCard key={video.video_id} video={video} isPlaying={props?.currentlyPlaying === video} onRemove={props.onRemoveVideo}/>
                            )
                        })}
                    </AccordionContent>
                </AccordionItem>
            </Accordion>
        </div>
    )
}
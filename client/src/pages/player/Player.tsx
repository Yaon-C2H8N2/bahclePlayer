import {useEffect, useState} from "react";
import {Button} from "@/components/ui/button.tsx";
import {fetchApi} from "@/lib/network.ts";
import {Playlist} from "./components/Playlist.tsx";
import ReactPlayer from "react-player";
import {SkipForward} from "lucide-react";
import {useToast} from "@/hooks/use-toast.ts";
import {ManualAddDialog} from "@/pages/player/components/ManualAddDialog.tsx";


export const Player = () => {
    const [playlist, setPlaylist] = useState<any>([])
    const [queue, setQueue] = useState<any>([])
    const [currentVideo, setCurrentVideo] = useState<any>(null)
    const [playlistIndex, setPlaylistIndex] = useState<number>(-1)
    const [openManualAdd, setOpenManualAdd] = useState<boolean>(false)
    const toast = useToast()

    const formatISODuration = (duration: string) => {
        const match = duration.match(/PT(?:(\d+)M)?(?:(\d+)S)?/);

        if (!match) {
            return '0:00';
        }
        const minutes = match[1] ? parseInt(match[1], 10) : 0;
        const seconds = match[2] ? parseInt(match[2], 10) : 0;

        const formattedSeconds = seconds.toString().padStart(2, '0');
        return `${minutes}:${formattedSeconds}`;
    }

    const addToQueue = (video: any) => {
        setQueue((prev: any[]) => [video, ...prev])
    }

    const addToPlaylist = (video: any) => {
        setPlaylist((prev: any[]) => [video, ...prev])
    }

    const initSocket = (playlistCallback: (video: any) => void, queueCallback: (video: any) => void) => {
           let socket = async (playlistCallback: (video: any) => void, queueCallback: (video: any) => void) => {
                const token = document.cookie.split(";").find(cookie => cookie.includes("token"))?.split("=")[1];
                const wsProtocol = window.location.protocol === "https:" ? "wss" : "ws";
                const ws = new WebSocket(`${wsProtocol}://${window.location.host}/api/player`);

                ws.onmessage = (event) => {
                    const data = JSON.parse(event.data)

                    if(data.welcome){
                        ws.send(JSON.stringify({
                            token: token
                        }))
                    } else if(data.error){
                        toast.toast({
                            title: "Error",
                            description: data.error,
                            variant: "destructive",
                        })
                        return
                    } else {
                        if(data.type === "QUEUE"){
                            queueCallback(data)
                        } else if(data.type === "PLAYLIST"){
                            playlistCallback(data)
                        }
                        toast.toast({
                            title: `A video as been added to the ${data.type === "QUEUE " ? "queue" : "playlist"}`,
                            description: data.title,
                            duration: 5000
                        })
                    }
                }

                return ws;
            };
            socket(playlistCallback, queueCallback)
    }

    useEffect(() => {
        if (!document.cookie.includes("token")){
            window.location.href = "/"
        }

        fetchApi("/api/playlist").then((res) => {
            if(res.status === 401){
                window.location.href = "/?error=token_expired"
            }
            return res.json();
        }).then((data) => {
            const videos: any[] = data.data ?? [];
            videos.forEach((video) => {
                video.duration = formatISODuration(video.duration)
            })
            const newQueue = videos.filter((video) => video.type === "QUEUE");
            const newPlaylist = videos.filter((video) => video.type === "PLAYLIST");

            setPlaylist(newPlaylist);
            setQueue(newQueue);
            if(newQueue.length > 0){
                setCurrentVideo(newQueue[0])
            } else {
                setCurrentVideo(newPlaylist[0])
                setPlaylistIndex(0)
            }
        })
        initSocket(addToPlaylist, addToQueue)
    }, [])

    const handleNextVideo = () => {
        if(currentVideo.type === "QUEUE") {
            const newQueue = queue.filter((video: any) => video.video_id !== currentVideo.video_id)
            removeVideo(currentVideo, true)
            if(newQueue.length > 0){
                setCurrentVideo(newQueue[0])
            } else {
                setPlaylistIndex((playlistIndex + 1)%playlist.length)
                setCurrentVideo(playlist[(playlistIndex + 1)%playlist.length])
            }
        } else {
            if (queue.length > 0) {
                setCurrentVideo(queue[0])
            } else {
                setPlaylistIndex((playlistIndex + 1)%playlist.length)
                setCurrentVideo(playlist[(playlistIndex + 1)%playlist.length])
            }
        }
    }

    const removeVideo = (video: any, auto: boolean) => {
        if (video.type === "QUEUE") {
            const newQueue = queue.filter((v: any) => v.video_id !== video.video_id)
            setQueue(newQueue)
            removeVideoFromApi(video, auto)
        } else {
            const newPlaylist = playlist.filter((v: any) => v.video_id !== video.video_id)
            setPlaylist(newPlaylist)
            removeVideoFromApi(video, auto)
        }
    }

    const removeVideoFromApi = (video: any, auto: boolean) => {
        fetchApi(`/api/playlist?video_id=${video.video_id}`, {method: "DELETE"})
            .then((res) => {
                return res.json()
            }
        ).then((data) => {
            if(data.status === "success" && !auto){
                toast.toast({
                    title: "Video removed",
                    description: `${video.title} as been removed from the playlist/queue`,
                    duration: 5000
                })
            }
        })
    }

    return (
        <div className={"flex flex-col items-center w-full h-[100vh]"}>
            <div className={"flex flex-row w-full p-14 gap-10"}>
                <div className={"flex w-2/3 justify-start items-center flex-col gap-2"}>
                    <div className={"w-full min-w-96 min-h-[60vh] p-5 justify-center items-center flex"}>
                        {playlist.length > 0 || queue.length > 0 ? (
                            <ReactPlayer
                                height={"100%"}
                                width={"100%"}
                                url={currentVideo?.url}
                                playing={true}
                                muted={true}
                                controls={true}
                                onEnded={() => {
                                    handleNextVideo()
                                }}
                            />
                        ) : (
                            <div>Playlist is empty</div>
                        )}
                    </div>
                    <div className={"flex flex-row items-center gap-2 justify-between w-2/3"}>
                        <div>{currentVideo?.title}</div>
                        <Button
                            variant={"outline"}
                            onClick={() => {
                                handleNextVideo()
                            }}
                        >
                            Skip <SkipForward/>
                        </Button>
                    </div>
                </div>
                <div className={"w-1/3 max-h-[80vh] min-h-[80vh] overflow-y-auto"}>
                    <Button onClick={() => setOpenManualAdd(true)}>Add track</Button>
                    <Playlist playlist={playlist} queue={queue} currentlyPlaying={currentVideo} onRemoveVideo={(video) => removeVideo(video, false)}/>
                </div>
            </div>
            <ManualAddDialog open={openManualAdd} onClose={() => setOpenManualAdd(false)} />
        </div>
    )
}
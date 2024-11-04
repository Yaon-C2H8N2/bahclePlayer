import {useEffect, useRef, useState} from "react";
import {Button} from "@/components/ui/button.tsx";
import {fetchApi} from "@/lib/network.ts";
import {Playlist} from "./components/Playlist.tsx";
import ReactPlayer from "react-player";
import {SkipForward} from "lucide-react";


export const Player = () => {
    const socketLoaded = useRef(false)
    const [playlist, setPlaylist] = useState<any>([])
    const [queue, setQueue] = useState<any>([])
    const [currentVideo, setCurrentVideo] = useState<any>(null)
    const [playlistIndex, setPlaylistIndex] = useState<number>(-1)

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

    const initSocket = () => {
        //weird websocket stuff to avoid strict mode
        let socket = null
        if(!socketLoaded.current){
            socketLoaded.current = true
            socket = async () => {
                const token = document.cookie.split(";").find(cookie => cookie.includes("token"))?.split("=")[1];
                const wsProtocol = window.location.protocol === "https:" ? "wss" : "ws";
                const ws = new WebSocket(`${wsProtocol}://localhost:8081/player?access_token=${token}`);

                ws.onmessage = (event) => {
                    const data = JSON.parse(event.data)
                    if(data.type === "QUEUE"){
                        setQueue([...queue, data.video])
                    } else if(data.type === "PLAYLIST"){
                        setPlaylist([...playlist, data.video])
                    }

                    //TODO : toaster success
                }

                return ws;
            };
            socket()
        }
    }

    useEffect(() => {
        if (!document.cookie.includes("token")){
            window.location.href = "/"
        }

        fetchApi("/api/playlist").then((res) => {
            return res.json();
        }).then((data) => {
            const videos: any[] = data.data;
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
        initSocket()
    }, [])

    const handleNextVideo = () => {
        if(currentVideo.type === "QUEUE") {
            const newQueue = queue.filter((video: any) => video.video_id !== currentVideo.video_id)
            setQueue(newQueue)
            removeVideo(currentVideo)
            if(newQueue.length > 0){
                setCurrentVideo(newQueue[0])
            } else {
                setPlaylistIndex((playlistIndex + 1)%playlist.length)
                setCurrentVideo(playlist[(playlistIndex + 1)%playlist.length])
            }
        } else {
            setPlaylistIndex((playlistIndex + 1)%playlist.length)
            setCurrentVideo(playlist[(playlistIndex + 1)%playlist.length])
        }
    }

    const removeVideo = (video: any) => {
        console.log(video)
        fetchApi("/api/playlist", {method: "DELETE", body: JSON.stringify({video_id: video.video_id})})
            .then((res) => {
                return res.json()
            }
        ).then((data) => {
            if(data.status === "success"){
                //TODO : toaster success
            }
        })
    }

    return (
        <div className={"flex flex-col items-center w-full h-[100vh]"}>
            <div className={"flex flex-row w-full p-14 gap-10"}>
                <div className={"flex w-2/3 justify-start items-center flex-col gap-2"}>
                    <div className={"w-full min-w-96 min-h-[60vh] p-5"}>
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
                    <Playlist playlist={playlist} queue={queue} currentlyPlaying={currentVideo} onRemoveVideo={(video) => removeVideo(video.video_id)}/>
                </div>
            </div>
        </div>
    )
}
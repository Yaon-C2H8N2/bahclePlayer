import {useEffect, useState} from "react";
import {useParams} from "react-router-dom";
import {VideoCard} from "@/pages/player/components/VideoCard.tsx";

export const CurrentlyPlaying = () => {
    const params = useParams() as { twitchId: string }
    const [currentTrack, setCurrentTrack] = useState<Video>();
    // const [userConfig, setUserConfig] = useState<any>();

    const initSocket = (twitchId: string) => {
        let socket = async (twitchId: string) => {
            const wsProtocol = window.location.protocol === "https:" ? "wss" : "ws";
            const ws = new WebSocket(`${wsProtocol}://${window.location.host}/api/overlays/events?twitch_id=${twitchId}&event_type=currently_playing`);

            ws.onmessage = (event) => {
                const data = JSON.parse(event.data) as Video
                setCurrentTrack(data)
            }

            return ws;
        };
        socket(twitchId)
    }

    // const fetchUserConfig = async (twitchId: string) => {
    //     const response = await fetch(`/api/overlays/user_config?twitch_id=${twitchId}`);
    //     if (response.ok) {
    //         const data = await response.json();
    //         return data;
    //     }
    // }

    useEffect(() => {
        initSocket(params.twitchId)
        // TODO : Add custom CSS to the overlay via the user config
        // fetchUserConfig(params.twitchId).then((data) => {
        //     setUserConfig(data)
        // })
    }, [])

    return (
        <>
            {currentTrack && (
                <VideoCard video={currentTrack}/>
            )}
        </>
    )
}
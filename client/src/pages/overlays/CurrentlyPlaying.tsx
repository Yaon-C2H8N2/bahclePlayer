import {useEffect, useState} from "react";
import {useParams} from "react-router-dom";
import {VideoCard} from "@/pages/player/components/VideoCard.tsx";

interface UserOverlaySettings {
    css?: string;
}

export const CurrentlyPlaying = () => {
    const params = useParams() as { twitchId: string }
    const [currentTrack, setCurrentTrack] = useState<Video>();
    const [userSettings, setUserSettings] = useState<UserOverlaySettings>({});

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

    const fetchUserSettings = async (twitchId: string) => {
        try {
            const response = await fetch(`/api/overlays/settings?twitch_id=${twitchId}&overlay_code=currently_playing`);
            if (response.ok) {
                const data = await response.json();
                const settings = JSON.parse(data.settings || '{}');
                setUserSettings(settings);
            }
        } catch (error) {
            console.error('Failed to fetch user settings:', error);
        }
    }

    useEffect(() => {
        initSocket(params.twitchId)
        fetchUserSettings(params.twitchId)
    }, [params.twitchId])

    // Apply custom CSS if provided
    useEffect(() => {
        if (userSettings.css) {
            const styleElement = document.createElement('style');
            styleElement.textContent = userSettings.css;
            styleElement.id = 'custom-overlay-css';
            document.head.appendChild(styleElement);

            return () => {
                const existingStyle = document.getElementById('custom-overlay-css');
                if (existingStyle) {
                    existingStyle.remove();
                }
            };
        }
    }, [userSettings]);

    return (
        <>
            {currentTrack && (
                <VideoCard video={currentTrack}/>
            )}
        </>
    )
}
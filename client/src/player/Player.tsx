import {useEffect, useRef, useState} from "react";
import {Button} from "@/components/ui/button.tsx";


export const Player = () => {
    const socketLoaded = useRef(false)
    const [messages, setMessages] = useState<any>([])

    useEffect(() => {
        if (!document.cookie.includes("token")){
            window.location.href = "/"
        }

        //weird websocket stuff to avoid strict mode
        let socket = null
        if(!socketLoaded.current){
            socketLoaded.current = true
            socket = async () => {
                const token = document.cookie.split(";").find(cookie => cookie.includes("token"))?.split("=")[1];
                const ws = new WebSocket(`ws://localhost:8081/player?access_token=${token}`);

                ws.onmessage = (event) => {
                    const data = JSON.parse(JSON.parse(event.data).message)
                    setMessages((prev: any) => [...prev, data])
                }

                return ws;
            };
            socket()
        }
    }, [])

    return (
        <div className={"flex flex-col items-center w-full h-[100vh]"}>
            <div className={"flex flex-row w-full p-3 items-center justify-end"}>
                <Button onClick={() => {
                    //todo: logout
                }}>Log out</Button>
            </div>
            <div className={"flex flex-col flex-1 w-full justify-center items-center"}>
                {messages.map((message: any, index: number) => {
                    return (
                        <div key={index}>
                            {message.payload.event.message.text}
                        </div>
                    )
                })}
            </div>
        </div>
    )
}
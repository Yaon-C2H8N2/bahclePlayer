import {Button} from "@/components/ui/button.tsx";
import {LoaderCircle, Twitch} from "lucide-react";
import {useEffect, useState} from "react";
import {fetchApi} from "@/lib/network.ts";

export const Home = () => {
    const [loading, setLoading] = useState(false)

    const handleLogin = async (token: string) => {
        const response = await fetchApi(`/api/login?access_token=${token}`)
        return response.json()
    }

    useEffect(() => {
        if (document.cookie.includes("token")) {
            const token = document.cookie.split(";").find(cookie => cookie.includes("token"))?.split("=")[1];
            handleLogin(token ?? "").then((response) => {
                if (response.error) {
                    throw new Error(response.error)
                } else {
                    window.location.href = "/player"
                }
            })
        } else {
            const fragments = document.location.hash.split("&")[0].split("=")
            fragments.forEach((fragment, index) => {
                if (fragment === "#access_token") {
                    setLoading(true)
                    handleLogin(fragments[index + 1]).then((response) => {
                        if (response.error) {
                            throw new Error(response.error)
                        } else {
                            window.location.href = "/player"
                        }
                    })
                }
            })
        }
    }, [])

    return (
        <div className={"flex flex-col gap-2 w-full h-[100vh] justify-center items-center"}>
            <div>Bahcle Player</div>
            <Button
                className={"bg-[#6441a5] hover:bg-[#6441a5] hover:brightness-75"}
                onClick={() => {
                    window.location.href = `https://id.twitch.tv/oauth2/authorize?response_type=token&client_id=${import.meta.env.VITE_TWITCH_CLIENT_ID}&redirect_uri=${import.meta.env.VITE_APP_URL}&scope=channel:moderate+whispers:edit+channel:read:redemptions+channel:manage:redemptions+channel:manage:polls+channel:read:polls+channel:bot`
                }}
            >
                {loading ? <LoaderCircle className={"animate-spin"}/> : <><Twitch/>Sign in with Twitch</>}
            </Button>
        </div>
    )
}
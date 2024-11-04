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
                    window.location.href = `https://id.twitch.tv/oauth2/authorize?response_type=token&client_id=${import.meta.env.VITE_TWITCH_CLIENT_ID}&redirect_uri=${import.meta.env.VITE_APP_URL}&scope=analytics:read:extensions+user:edit+user:read:email+clips:edit+bits:read+analytics:read:games+user:edit:broadcast+user:read:broadcast+chat:read+chat:edit+channel:moderate+channel:read:subscriptions+whispers:read+whispers:edit+moderation:read+channel:read:redemptions+channel:edit:commercial+channel:read:hype_train+channel:read:stream_key+channel:manage:extensions+channel:manage:broadcast+user:edit:follows+channel:manage:redemptions+channel:read:editors+channel:manage:videos+user:read:blocked_users+user:manage:blocked_users+user:read:subscriptions+user:read:follows+channel:manage:polls+channel:manage:predictions+channel:read:polls+channel:read:predictions+moderator:manage:automod+channel:manage:schedule+channel:read:goals+moderator:read:automod_settings+moderator:manage:automod_settings+moderator:manage:banned_users+moderator:read:blocked_terms+moderator:manage:blocked_terms+moderator:read:chat_settings+moderator:manage:chat_settings+channel:manage:raids+moderator:manage:announcements+moderator:manage:chat_messages+user:manage:chat_color+channel:manage:moderators+channel:read:vips+channel:manage:vips+user:manage:whispers+channel:read:charity+moderator:read:chatters+moderator:read:shield_mode+moderator:manage:shield_mode+moderator:read:shoutouts+moderator:manage:shoutouts+moderator:read:followers+channel:read:guest_star+channel:manage:guest_star+moderator:read:guest_star+moderator:manage:guest_star+channel:bot+user:bot+user:read:chat+channel:manage:ads+channel:read:ads+user:read:moderated_channels+user:write:chat+user:read:emotes+moderator:read:unban_requests+moderator:manage:unban_requests+moderator:read:suspicious_users+moderator:manage:warnings`
                }}
            >
                {loading ? <LoaderCircle className={"animate-spin"}/> : <><Twitch/>Sign in with Twitch</>}
            </Button>
        </div>
    )
}
import {Button} from "@/components/ui/button.tsx";
import {Twitch} from "lucide-react";

export const Home = () => {
    return (
        <div className={"flex flex-col gap-2 w-full h-[100vh] justify-center items-center"}>
            <div>Bahcle Player</div>
            <Button className={"bg-[#6441a5] hover:bg-[#6441a5] hover:brightness-75"}><Twitch/>Sign in with Twitch</Button>
        </div>
    )
}
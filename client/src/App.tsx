import {BrowserRouter, Route, Routes} from "react-router-dom";
import {Home} from "./pages/home/Home.tsx";
import {Player} from "./pages/player/Player.tsx";
import {Layout} from "@/components/Layout.tsx";
import {CurrentlyPlaying} from "@/pages/overlays/CurrentlyPlaying.tsx";

export const App = () => {

    return (
        <>
            <BrowserRouter>
                <Routes>
                    <Route element={<Layout/>}>
                        <Route path={"/player/*"} element={<Player/>}/>
                    </Route>
                    <Route path={"/overlay/:twitchId/*"}>
                        <Route path={"currently_playing"} element={<CurrentlyPlaying/>}/>
                        <Route path={"playlist_and_queue"} element={<div>Playlist and queue</div>}/>
                        <Route path={"add_events"} element={<div>Add events</div>}/>
                    </Route>
                    <Route path={"/*"} element={<Home/>}/>
                </Routes>
            </BrowserRouter>
        </>
    )
}
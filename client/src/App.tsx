import {BrowserRouter, Route, Routes} from "react-router-dom";
import {Home} from "./pages/home/Home.tsx";
import {Player} from "./pages/player/Player.tsx";
import {Layout} from "@/components/Layout.tsx";

export const App = () => {

    return (
        <>
            <BrowserRouter>
                <Routes>
                    <Route element={<Layout/>}>
                        <Route path={"/player/*"} element={<Player/>}/>
                    </Route>
                    <Route path={"/*"} element={<Home/>}/>
                </Routes>
            </BrowserRouter>
        </>
    )
}
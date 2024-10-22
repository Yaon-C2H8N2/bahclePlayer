import {BrowserRouter, Route, Routes} from "react-router-dom";
import {Home} from "./home/Home.tsx";
import {Player} from "./player/Player.tsx";

export const App = () => {

    return (
        <>
            <BrowserRouter>
                <Routes>
                    <Route path="/player/*" element={<Player/>}/>
                    <Route path="/*" element={<Home/>}/>
                </Routes>
            </BrowserRouter>
        </>
    )
}
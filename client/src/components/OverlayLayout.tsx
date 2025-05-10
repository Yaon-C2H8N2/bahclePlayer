import {Outlet} from "react-router-dom";

export const OverlayLayout = () => {
    return (
        <div className={"bg-transparent"}>
            <Outlet/>
        </div>
    )
}

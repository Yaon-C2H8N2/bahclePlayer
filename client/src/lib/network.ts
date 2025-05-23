export const fetchApi = async (url: string, options: RequestInit = {}) => {
    const token = document.cookie.split(";").find(cookie => cookie.includes("token"))?.split("=")[1];
    if (token) {
        options.headers = {
            ...options.headers,
            "Authorization": `Bearer ${token}`
        }
    }

    return await fetch(url, options);
}
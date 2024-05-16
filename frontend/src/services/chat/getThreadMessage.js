
import axiosInstance from "../configs";

export default async function getThreadMessage(threadId) {
    
    try {
        let res = await axiosInstance.get(`/thread/${threadId}/messages`)
        
        res = res.data.meta.code === 200 ? res.data.data : { error: res.data.meta.message }
        console.log(res)
        
        return res;
    }
    catch (err) {
        if (err.response) {
            console.log(err.response.data.message);
            return { error: err.response.data.message };
        }
    }
}
import { useNavigate, useParams } from "react-router-dom"
import { API_URL } from "./App"

//确定页面
export const ConfirmationPage = () => {
    //从URL中读取参数
    const { token = '' } = useParams()
    //重定向函数
    const redirect = useNavigate()
    //处理点击事件,发送网络请求
    const handleConfirm = async () => {
        const response = await fetch(`${API_URL}/users/activate/${token}`,{
            method: "PUT"
        })
        //检查是否请求成功
        if(response.ok){
            //重定向到 "/"路径
            redirect('/')
        }else {
            //跳转到error页面
            //TODO
            alert("Failed to confirm token")
        }
    }
    return (
        <>
            <h1>Confirmation</h1>
            <button onClick={handleConfirm}>Click to confirm </button>
        </>
    )
}
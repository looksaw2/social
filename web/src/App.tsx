import './App.css'
//定义一些环境变量，需要传入的
export const API_URL =  import.meta.env.VITE_API_URL || "http://localhost:8080/v1"

function App() {
  return (
    <>
      <div>App home screen</div>
    </>
  )
}

export default App

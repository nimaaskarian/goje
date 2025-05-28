import { render } from 'preact';
import { useEffect, useState } from 'preact/hooks';

import './style.css';

export function App() {
  const [timer, setTimer] = useState()
  const postTimer = () => {
    let xhr = new XMLHttpRequest();
    xhr.open("POST", '/api/timer', true);
    xhr.setRequestHeader("Content-Type", "application/json; charset=UTF-8")
    xhr.responseType = 'json'
    xhr.onload = () => {
      if (xhr.status == 200) {
        setTimer(xhr.response)
      }
    }
    xhr.send(JSON.stringify(timer));
  }
  useEffect(()=> {
    const sse = new EventSource("/api/timer/stream")
    console.log("oh my")
    sse.onopen = e => {
      console.log("open", e)
    }
    sse.addEventListener("timer", (e)=> {
      console.log(JSON.parse(e.data))
      setTimer(JSON.parse(e.data))
    })
    sse.onerror = e => {
      console.log(e)
    }
    return () => {
      sse.close()
    }
  }, [])
  if (timer) {
    return (
    <div>
      <Timer duration={timer.Duration}/>
      <input type="button" value="pause" onClick={()=> {
          timer.Paused = !timer.Paused;
          postTimer()
        }} />
    </div>
    );
  }
}

function Timer(props) {
    let seconds = props.duration/1E9;
    let minutes = (seconds/60>>0)
    seconds = seconds%60
    let hours = (minutes/60>>0)
    minutes = minutes%60
    return (
      <div>
        <p>
          {String(minutes).padStart(2,'0')}:{String(seconds).padStart(2,'0')}
        </p>
      </div>
    );
}

render(<App />, document.getElementById('app'));

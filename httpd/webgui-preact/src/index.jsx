import { render } from 'preact';
import { useEffect, useState } from 'preact/hooks';

import './style.css';

const pause_icon = <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" class="size-8">
  <path d="M5.75 3a.75.75 0 0 0-.75.75v12.5c0 .414.336.75.75.75h1.5a.75.75 0 0 0 .75-.75V3.75A.75.75 0 0 0 7.25 3h-1.5ZM12.75 3a.75.75 0 0 0-.75.75v12.5c0 .414.336.75.75.75h1.5a.75.75 0 0 0 .75-.75V3.75a.75.75 0 0 0-.75-.75h-1.5Z" />
</svg>
const play_icon = <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" class="size-8">
  <path d="M6.3 2.84A1.5 1.5 0 0 0 4 4.11v11.78a1.5 1.5 0 0 0 2.3 1.27l9.344-5.891a1.5 1.5 0 0 0 0-2.538L6.3 2.841Z" />
</svg>

const prev_icon = <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" class="size-8">
  <path fillRule="evenodd" d="M11.78 5.22a.75.75 0 0 1 0 1.06L8.06 10l3.72 3.72a.75.75 0 1 1-1.06 1.06l-4.25-4.25a.75.75 0 0 1 0-1.06l4.25-4.25a.75.75 0 0 1 1.06 0Z" clipRule="evenodd" />
</svg>
const next_icon = <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" class="size-8">
  <path fillRule="evenodd" d="M8.22 5.22a.75.75 0 0 1 1.06 0l4.25 4.25a.75.75 0 0 1 0 1.06l-4.25 4.25a.75.75 0 0 1-1.06-1.06L11.94 10 8.22 6.28a.75.75 0 0 1 0-1.06Z" clipRule="evenodd" />
</svg>



export function App() {
  const [timer, setTimer] = useState()
  const postTimer = (endpoint) => {
    let xhr = new XMLHttpRequest();
    xhr.open("POST", '/api/timer' + (endpoint ? "/" + endpoint : ""), true);
    xhr.setRequestHeader("Content-Type", "application/json; charset=UTF-8")
    xhr.responseType = 'json'
    xhr.send(JSON.stringify(timer));
  }
  useEffect(() => {
    const sse = new EventSource("/api/timer/stream")
    sse.addEventListener("timer", (e) => {
      setTimer(JSON.parse(e.data))
    })
    return () => {
      sse.close()
    }
  }, [])
  if (timer) {
    let mode = ""
    switch (timer.Mode) {
      case 0:
        mode = "Pomodoro"
        break;
      case 1:
        mode = "Short Break"
        break;
      case 2:
        mode = "Long Break"
        break;
    }
    return (
      <div class="min-h-screen flex flex-col justify-center items-center bg-zinc-300 text-zinc-900 dark:text-white dark:bg-zinc-900">
        <div class="w-40 text-center flex flex-col gap-3">
          <div class="dark:bg-zinc-800 bg-white rounded-lg p-4 shadow-md">
            {mode}
            <Timer duration={timer.Duration} />
          </div>
          <div class="flex justify-center gap-3 text-2xl">
            <button class="cursor-pointer" onClick={() => { postTimer("prevmode") }}>
              {prev_icon}
            </button>
            <button onClick={() => {
              timer.Paused = !timer.Paused;
              postTimer()
            }} class="cursor-pointer p-2 rounded-full bg-white dark:bg-zinc-800">
              {timer.Paused ? play_icon : pause_icon}
            </button>
            <button class="cursor-pointer" onClick={() => { postTimer("nextmode") }}>
              {next_icon}
            </button>
          </div>
          <div class="flex justify-center gap-2">
            <button class="cursor-pointer" onClick={() => { timer.SessionCount--; postTimer(); }}>
              <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" class="size-5">
                <path fillRule="evenodd" d="M4 10a.75.75 0 0 1 .75-.75h10.5a.75.75 0 0 1 0 1.5H4.75A.75.75 0 0 1 4 10Z" clipRule="evenodd" />
              </svg>
            </button>
            {timer.SessionCount}
            <button class="cursor-pointer" onClick={() => { timer.SessionCount++; postTimer(); }}>
              <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" class="size-5">
                <path d="M10.75 4.75a.75.75 0 0 0-1.5 0v4.5h-4.5a.75.75 0 0 0 0 1.5h4.5v4.5a.75.75 0 0 0 1.5 0v-4.5h4.5a.75.75 0 0 0 0-1.5h-4.5v-4.5Z" />
              </svg>
            </button>

          </div>
        </div>
      </div >
    );
  }
}

function Timer(props) {

  let seconds = props.duration / 1E9;
  let minutes = (seconds / 60 >> 0)
  seconds = seconds % 60
  let hours = (minutes / 60 >> 0)
  minutes = minutes % 60
  if (hours !== 0) {
    return (
      <div class={props.class}>
        {String(hours).padStart(2, '0')}:{String(minutes).padStart(2, '0')}:{String(seconds).padStart(2, '0')}
      </div>
    );
  } else {
    return (
      <div class={props.class}>
        {String(minutes).padStart(2, '0')}:{String(seconds).padStart(2, '0')}
      </div>
    );
  }
}

render(<App />, document.getElementById('app'));

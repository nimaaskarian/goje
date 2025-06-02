import { render } from 'preact';
import { useMemo, useState } from 'preact/hooks';
import { Settings } from "./settings"
import { Button } from "./utils"
import { postTimer, timerModeString } from "./timer"

import './style.css';

/**
 * @typedef {import("./timer.js").Timer} Timer
 */

export function App() {
  /** @type {[Timer, (timer: Timer) => void]} */
  const [timer, setTimer] = useState(undefined)
  const [settings, setSettings] = useState(false)
  const sse = useMemo(() => {
    const sse = new EventSource("/api/timer/stream")
    sse.addEventListener("change", (e) => {
      setTimer(JSON.parse(e.data))
    })
    sse.addEventListener("pause", (e) => {
      setTimer(JSON.parse(e.data))
    })
    sse.onerror = _ => {
      setTimer(null)
    }
    window.addEventListener("beforeunload", () => {
      sse.close()
    })
    return sse;
  }, [])
  if (timer) {
    return (
      <div class={"h-full flex flex-col justify-center items-center bg-zinc-200 text-zinc-900 dark:text-white dark:bg-zinc-900" + (settings ? " overflow-hidden" : "")}>
        <Settings onClose={() => setSettings(false)} timer={timer} hidden={!settings} sse={sse} />
        <button title="open settings" aria-label="open settings" onClick={() => setSettings(true)} class="absolute top-4 right-4 p-2 rounded dark:bg-zinc-800 bg-white shadow-sm hover:shadow-md transition ease-in-out duration-150 hover:text-zinc-600 hover:dark:text-zinc-300 cursor-pointer z-0">
          {cog_icon}
        </button>
        <div class="min-w-60 text-center flex flex-col gap-4">
          <div class="dark:bg-zinc-800 bg-white rounded-lg p-4 flex gap-4 flex-col shadow-sm hover:shadow-md transition ease-in-out duration-150">
            <ModeSelection timer={timer} />
            <div class="flex flex-row justify-center gap-2">
              <Button title="-1 finished sessions" onClick={() => { timer.FinishedSessions--; postTimer(timer); }}>
                {minus_icon}
              </Button>
              {timer.FinishedSessions}/{timer.Config.Sessions}
              <Button title="+1 finished sessions" onClick={() => { timer.FinishedSessions++; postTimer(timer); }}>
                {plus_icon}
              </Button>
            </div>

            <TimerCircle timer={timer} />
            <div class="flex justify-center gap-4">
              <Button title="Previous mode" onClick={() => { postTimer(timer, "/prevmode") }}>
                {prev_icon}
              </Button>
              <Button title={`${timer.Paused ? "Resume" : "Pause"} timer`} onClick={() => { postTimer(timer, "/pause") }}>
                {timer.Paused ? play_icon : pause_icon}
              </Button>
              <Button title="Next mode" onClick={() => { postTimer(timer, "/nextmode") }}>
                {next_icon}
              </Button>
            </div>
          </div>
        </div>
      </div>
    );
  }
  if (timer === null) {
    return (
      <div class="h-full flex flex-col justify-center items-center bg-zinc-200 text-zinc-900 dark:text-white dark:bg-zinc-900">
        Goje isn't running :(
      </div>
    )
  }
}

function TimerCircle(p) {
  const progress = useMemo(() => {
    if (p.timer) {
      const total_duration = p.timer.Config.Duration[p.timer.Mode]
      return `${((total_duration - p.timer.Duration) / total_duration) * 100}%`
    }
  }, [p.timer])
  return (
    <div class="circle flex justify-center items-center" id="timer-circle" style={{ "--progress": progress }}>
      <div class="inner-circle flex gap-2 justify-center items-center bg-white dark:bg-zinc-800">
        <Button title="Reset timer" onClick={() => { postTimer(p.timer, "/reset") }}>
          {restart_icon}
        </Button>
        <Timer timer={p.timer} />
      </div>
    </div>
  );
}

function ModeSelection(p) {
  const modeOptions = useMemo(() => {
    return [0, 1, 2].map((mode) =>
      <option value={mode} class="checked:dark:bg-zinc-700 checked:bg-zinc-300 hover:bg-zinc-300">
        {timerModeString(mode)}
      </option>
    )
  }, [])

  return (
    <select aria-label="Timer mode" title="Timer mode" value={p.timer.Mode} class="dark:bg-zinc-900 bg-zinc-200 p-2 rounded" onChange={(e) => {
      p.timer.Mode = parseInt(e.target.value);
      postTimer(p.timer);
    }}>
      {modeOptions}
    </select>
  );
}

function Timer(props) {
  const [fraction, seconds, minutes, hours] = useMemo(() => {
    let fraction = ""
    let fraclen = 0
    if (props.timer.Config.DurationPerTick < 1E9) {
      fraclen = Math.log10(1E9 / props.timer.Config.DurationPerTick + 1) >> 0
      const fraction_value = (props.timer.Duration % 1E9) / props.timer.Config.DurationPerTick
      fraction = `.${String(fraction_value).padStart(fraclen, '0')}`
    }
    let seconds = (props.timer.Duration / 1E9 >> 0);
    let minutes = (seconds / 60 >> 0)
    seconds = seconds % 60
    let hours_value = (minutes / 60 >> 0)
    let hours = ""
    if (hours_value !== 0) {
      hours = String(hours_value).padStart(2, '0') + ':'
    }
    minutes = minutes % 60
    return [fraction, seconds, minutes, hours]
  }, [props.timer.Duration, props.timer.Config.DurationPerTick])

  return (
    <div class="text-2xl font-bold">
      {hours}{String(minutes).padStart(2, '0')}:{String(seconds).padStart(2, '0')}{fraction}
    </div>
  );
}

const pause_icon = <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={3} stroke="currentColor" className="size-10">
  <path strokeLinecap="round" strokeLinejoin="round" d="M15.75 5.25v13.5m-7.5-13.5v13.5" />
</svg>

const play_icon = <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={3} stroke="currentColor" className="size-10">
  <path strokeLinecap="round" strokeLinejoin="round" d="M5.25 5.653c0-.856.917-1.398 1.667-.986l11.54 6.347a1.125 1.125 0 0 1 0 1.972l-11.54 6.347a1.125 1.125 0 0 1-1.667-.986V5.653Z" />
</svg>

const next_icon = <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={30 / 7} stroke="currentColor" className="size-7">
  <path strokeLinecap="round" strokeLinejoin="round" d="M3 8.689c0-.864.933-1.406 1.683-.977l7.108 4.061a1.125 1.125 0 0 1 0 1.954l-7.108 4.061A1.125 1.125 0 0 1 3 16.811V8.69ZM12.75 8.689c0-.864.933-1.406 1.683-.977l7.108 4.061a1.125 1.125 0 0 1 0 1.954l-7.108 4.061a1.125 1.125 0 0 1-1.683-.977V8.69Z" />
</svg>

const prev_icon = <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={30 / 7} stroke="currentColor" className="size-7">
  <path strokeLinecap="round" strokeLinejoin="round" d="M21 16.811c0 .864-.933 1.406-1.683.977l-7.108-4.061a1.125 1.125 0 0 1 0-1.954l7.108-4.061A1.125 1.125 0 0 1 21 8.689v8.122ZM11.25 16.811c0 .864-.933 1.406-1.683.977l-7.108-4.061a1.125 1.125 0 0 1 0-1.954l7.108-4.061a1.125 1.125 0 0 1 1.683.977v8.122Z" />
</svg>

const cog_icon = <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="size-6">
  <path strokeLinecap="round" strokeLinejoin="round" d="M9.594 3.94c.09-.542.56-.94 1.11-.94h2.593c.55 0 1.02.398 1.11.94l.213 1.281c.063.374.313.686.645.87.074.04.147.083.22.127.325.196.72.257 1.075.124l1.217-.456a1.125 1.125 0 0 1 1.37.49l1.296 2.247a1.125 1.125 0 0 1-.26 1.431l-1.003.827c-.293.241-.438.613-.43.992a7.723 7.723 0 0 1 0 .255c-.008.378.137.75.43.991l1.004.827c.424.35.534.955.26 1.43l-1.298 2.247a1.125 1.125 0 0 1-1.369.491l-1.217-.456c-.355-.133-.75-.072-1.076.124a6.47 6.47 0 0 1-.22.128c-.331.183-.581.495-.644.869l-.213 1.281c-.09.543-.56.94-1.11.94h-2.594c-.55 0-1.019-.398-1.11-.94l-.213-1.281c-.062-.374-.312-.686-.644-.87a6.52 6.52 0 0 1-.22-.127c-.325-.196-.72-.257-1.076-.124l-1.217.456a1.125 1.125 0 0 1-1.369-.49l-1.297-2.247a1.125 1.125 0 0 1 .26-1.431l1.004-.827c.292-.24.437-.613.43-.991a6.932 6.932 0 0 1 0-.255c.007-.38-.138-.751-.43-.992l-1.004-.827a1.125 1.125 0 0 1-.26-1.43l1.297-2.247a1.125 1.125 0 0 1 1.37-.491l1.216.456c.356.133.751.072 1.076-.124.072-.044.146-.086.22-.128.332-.183.582-.495.644-.869l.214-1.28Z" />
  <path strokeLinecap="round" strokeLinejoin="round" d="M15 12a3 3 0 1 1-6 0 3 3 0 0 1 6 0Z" />
</svg>

const minus_icon = <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" class="size-5">
  <path fillRule="evenodd" d="M4 10a.75.75 0 0 1 .75-.75h10.5a.75.75 0 0 1 0 1.5H4.75A.75.75 0 0 1 4 10Z" clipRule="evenodd" />
</svg>
const plus_icon = <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" class="size-5">
  <path d="M10.75 4.75a.75.75 0 0 0-1.5 0v4.5h-4.5a.75.75 0 0 0 0 1.5h4.5v4.5a.75.75 0 0 0 1.5 0v-4.5h4.5a.75.75 0 0 0 0-1.5h-4.5v-4.5Z" />
</svg>

const restart_icon = <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={6} stroke="currentColor" className="size-5">
  <path strokeLinecap="round" strokeLinejoin="round" d="M16.023 9.348h4.992v-.001M2.985 19.644v-4.992m0 0h4.992m-4.993 0 3.181 3.183a8.25 8.25 0 0 0 13.803-3.7M4.031 9.865a8.25 8.25 0 0 1 13.803-3.7l3.181 3.182m0-4.991v4.99" />
</svg>

render(<App />, document.getElementById('app'));

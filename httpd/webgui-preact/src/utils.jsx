export function Radio(p) {
  return (
    <div class="flex items-center gap-2 box-border">
      <input defaultChecked={p.checked} onChange={p.onChange} id={p.id} type="checkbox"
        class="appearance-none
        cursor-pointer
        checked:dark:bg-zinc-500
        dark:border-zinc-500
        checked:bg-zinc-500
        border-zinc-500
        border-2
        w-3.5
        aspect-square
        rounded-full
    disabled:border-gray-400"
      />
      <label class="cursor-pointer" htmlFor={p.id}>{p.children}</label>
    </div>
  )
}

// a button with customized styles that uses its children like a good normal element.
// also sets the title prop as its aria-label as well. a11y yay
export function Button(props) {
  return (
    <button id={props.id} title={props.title} aria-label={props.title} onClick={props.onClick} class="cursor-pointer transition hover:text-zinc-600 hover:dark:text-zinc-300 ease-out">
      {props.children}
    </button>
  );
}

const ns_in_ms = 1_000_000
const ns_in_s = 1_000_000_000
const ns_in_m = ns_in_s * 60
const ns_in_h = ns_in_m * 60

/**
 * @param {number} nanoseconds
 */
export function formatDuration(nanoseconds) {
  let hours = (nanoseconds / ns_in_h) >> 0
  nanoseconds %= ns_in_h
  let minutes = (nanoseconds / ns_in_m) >> 0
  nanoseconds %= ns_in_m
  let seconds = (nanoseconds / ns_in_s) >> 0
  nanoseconds %= ns_in_s
  let miliseconds = (nanoseconds / ns_in_ms) >> 0
  nanoseconds %= ns_in_ms
  return [[hours, "h"], [minutes, "m"], [seconds, "s"], [miliseconds, "ms"], [nanoseconds, "ns"]].filter((e) => e[0]).flat().join("")
}

/**
 * @param {string} duration 
 * @returns {number} duration as nanoseconds
 */
export function parseDuration(duration) {
  let nanoseconds = 0
  let nanoseconds_str = duration.match(/(\d+)ns/)
  if (nanoseconds_str) {
    nanoseconds = parseInt(nanoseconds_str[1])
  }
  let miliseconds = duration.match(/(\d+)ms/);
  let seconds = duration.match(/(\d+)s/);
  let hours = duration.match(/(\d+)h/);
  let minutes = duration.match(/(\d+)m(?!s)/);
  if (hours) {
    nanoseconds += parseInt(hours[1]) * ns_in_h;
  } if (minutes) {
    nanoseconds += parseInt(minutes[1]) * ns_in_m;
  } if (seconds) {
    nanoseconds += parseInt(seconds[1]) * ns_in_s
  } if (miliseconds) {
    nanoseconds += parseInt(miliseconds[1]) * ns_in_ms
  }
  return nanoseconds
}


export function sendNotification(body) {
  let notif = {
      body, 
      icon: window.location.origin+"/assets/goje-512x512.png",
    }
  console.log(notif)
    const n = new Notification("Goje", notif)
    document.addEventListener("visibilitychange", () => {
      if (document.visibilityState === "visible") {
        n.close();
      }
    });
}

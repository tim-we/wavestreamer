import type { FunctionComponent } from "preact";
import type { HistoryEntry } from "../wavestreamer-api";

type HistoryProps = {
  data: HistoryEntry[];
};

const History: FunctionComponent<HistoryProps> = function History({ data }) {
  const history = data.slice().reverse();

  return (
    <section id="history">
      <div class="title">Recent history:</div>
      <table id="history-clips">
        <tbody>
          {history.map((entry) => {
            const title = entry.userScheduled ? (
              <i>{entry.title}</i>
            ) : (
              entry.title
            );
            return (
              <tr key={entry.start + entry.title} class="clip">
                <td>{dateToLocalTime(entry.start)}</td>
                <td>
                  {entry.skipped ? <s title="skipped">{title}</s> : title}
                </td>
              </tr>
            );
          })}
        </tbody>
      </table>
    </section>
  );
};

export default History;

function dateToLocalTime(time: string | undefined): string {
  if (!time) {
    return "";
  }

  // Example: 2025-04-21T10:41:00.236652254+02:00
  // Remove nanoseconds (not supported by JS Date)
  const date = new Date(time.replace(/\.\d+/, ""));

  return date.toLocaleTimeString([], {
    hour: "2-digit",
    minute: "2-digit",
  });
}

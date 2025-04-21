import { Component } from "preact";
import { HistoryEntry } from "../wavestreamer";

type HistoryProps = {
    data: HistoryEntry[];
};

export default class History extends Component<HistoryProps> {
    public render() {
        const history = this.props.data.slice().reverse();
        return (
            <div id="history">
                <div className="title">Recent history:</div>
                <div id="history-clips">
                    {history.map((clip) => {
                        const content = (
                            <>
                                {`${dateToLocalTime(clip.start)} `}
                                {clip.userScheduled ? (
                                    <i>{clip.title}</i>
                                ) : (
                                    clip.title
                                )}
                            </>
                        );

                        return (
                            <div
                                className={"clip"}
                                key={clip.start + clip.title}
                            >
                                {clip.skipped ? (
                                    <s title="skipped">{content}</s>
                                ) : (
                                    content
                                )}
                            </div>
                        );
                    })}
                </div>
            </div>
        );
    }
}

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

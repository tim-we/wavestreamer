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
                                {`${clip.start} `}
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

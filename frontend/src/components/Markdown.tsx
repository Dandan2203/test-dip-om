import { Fragment, type ReactNode } from "react";

function renderInline(text: string): ReactNode[] {
  return text.split(/(\*\*[^*]+\*\*)/g).map((part, i) => {
    const bold = /^\*\*([^*]+)\*\*$/.exec(part);
    return bold ? <strong key={i}>{bold[1]}</strong> : <Fragment key={i}>{part}</Fragment>;
  });
}

const BULLET = /^\s*[-*•]\s+(.*)$/;

export function Markdown({ content }: { content: string }) {
  const lines = content.replace(/\r\n/g, "\n").split("\n");
  const blocks: ReactNode[] = [];
  let para: string[] = [];
  let list: string[] = [];

  const flushPara = () => {
    if (!para.length) return;
    const cur = para;
    blocks.push(
      <p key={`p${blocks.length}`}>
        {cur.map((ln, i) => (
          <Fragment key={i}>
            {i > 0 && <br />}
            {renderInline(ln)}
          </Fragment>
        ))}
      </p>,
    );
    para = [];
  };

  const flushList = () => {
    if (!list.length) return;
    const cur = list;
    blocks.push(
      <ul key={`u${blocks.length}`} className="list-disc space-y-0.5 pl-4">
        {cur.map((ln, i) => (
          <li key={i}>{renderInline(ln)}</li>
        ))}
      </ul>,
    );
    list = [];
  };

  for (const line of lines) {
    const bullet = BULLET.exec(line);
    if (bullet) {
      flushPara();
      list.push(bullet[1]);
    } else if (line.trim() === "") {
      flushPara();
      flushList();
    } else {
      flushList();
      para.push(line);
    }
  }
  flushPara();
  flushList();

  return <div className="space-y-2 break-words">{blocks}</div>;
}

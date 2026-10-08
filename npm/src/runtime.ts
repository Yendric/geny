import { createElement, type ComponentType } from "react";
import { createRoot, type Root } from "react-dom/client";

type Props = Record<string, unknown>;
type Loader = () => Promise<{ default: ComponentType<Props> }>;

function islandName(path: string, prefix: string): string {
  return path.slice(prefix.length).replace(/\.[jt]sx$/, "");
}

const roots = new Map<HTMLElement, Root>();
const seen = new WeakSet<HTMLElement>();

async function mount(el: HTMLElement, loader: Loader): Promise<void> {
  const props: Props = JSON.parse(el.getAttribute("props") ?? "{}");
  const { default: component } = await loader();
  if (!el.isConnected) return;
  const root = createRoot(el);
  roots.set(el, root);
  root.render(createElement(component, props));
}

function whenVisible(el: HTMLElement, callback: () => void): void {
  const observer = new IntersectionObserver((entries) => {
    if (entries.some((entry) => entry.isIntersecting)) {
      observer.disconnect();
      callback();
    }
  });
  observer.observe(el);
}

export function start(modules: Record<string, Loader>, prefix: string): void {
  const loaders = new Map<string, Loader>();
  for (const [path, loader] of Object.entries(modules)) {
    loaders.set(islandName(path, prefix), loader);
  }

  const hydrate = (el: HTMLElement) => {
    if (seen.has(el)) return;
    seen.add(el);

    const name = el.getAttribute("component") ?? "";
    const loader = loaders.get(name);
    if (loader === undefined) {
      console.error(`geny: unknown island ${name}`);
      return;
    }

    const run = () => {
      mount(el, loader).catch((error: unknown) => console.error(error));
    };

    if (el.getAttribute("client") === "visible") {
      whenVisible(el, run);
    } else {
      run();
    }
  };

  const scan = (root: ParentNode) => {
    root.querySelectorAll<HTMLElement>("geny-island").forEach(hydrate);
  };

  scan(document);

  new MutationObserver((records) => {
    let removed = false;
    for (const record of records) {
      removed ||= record.removedNodes.length > 0;
      for (const node of record.addedNodes) {
        if (!(node instanceof HTMLElement)) continue;
        if (node.parentElement?.closest("geny-island")) continue;
        if (node.matches("geny-island")) hydrate(node);
        else scan(node);
      }
    }
    if (removed) {
      for (const [el, root] of roots) {
        if (!el.isConnected) {
          root.unmount();
          roots.delete(el);
        }
      }
    }
  }).observe(document.documentElement, { childList: true, subtree: true });
}

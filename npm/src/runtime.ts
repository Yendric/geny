import { createElement, type ComponentType } from "react";
import { createRoot } from "react-dom/client";

type Props = Record<string, unknown>;
type Loader = () => Promise<{ default: ComponentType<Props> }>;

function islandName(path: string, prefix: string): string {
  return path.slice(prefix.length).replace(/\.[jt]sx$/, "");
}

async function mount(el: HTMLElement, loader: Loader): Promise<void> {
  const props: Props = JSON.parse(el.getAttribute("props") ?? "{}");
  const { default: component } = await loader();
  createRoot(el).render(createElement(component, props));
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

  for (const el of document.querySelectorAll<HTMLElement>("geny-island")) {
    const name = el.getAttribute("component") ?? "";
    const loader = loaders.get(name);
    if (loader === undefined) {
      console.error(`geny: unknown island ${name}`);
      continue;
    }

    const hydrate = () => {
      mount(el, loader).catch((error: unknown) => console.error(error));
    };

    if (el.getAttribute("client") === "visible") {
      whenVisible(el, hydrate);
    } else {
      hydrate();
    }
  }
}

import assert from "node:assert/strict";
import test from "node:test";
import { createDesktopBridge } from "../app/lib/desktop.ts";

function mockInterop() {
  const calls = [];
  let eventHandler;
  const unsubscribe = () => {};
  return {
    calls,
    unsubscribe,
    emit(payload) {
      eventHandler?.({ event: "racore://daemon-exit", id: 1, payload });
    },
    interop: {
      async invoke(command, args) {
        calls.push({ command, args });
        return { command, args };
      },
      async listen(event, handler) {
        calls.push({ event });
        eventHandler = handler;
        return unsubscribe;
      },
    },
  };
}

test("desktop bridge maps command names and payload envelopes", async () => {
  const mock = mockInterop();
  const bridge = createDesktopBridge(mock.interop);
  const request = { path: "/v1/chat", method: "POST", body: { text: "hi" } };

  await bridge.status();
  await bridge.api(request);
  await bridge.platform();
  await bridge.openBrowser("https://racore.xyz");
  await bridge.openExternal("https://racore.xyz/docs");

  assert.deepEqual(mock.calls, [
    { command: "daemon_status", args: undefined },
    { command: "daemon_request", args: { request } },
    { command: "platform_info", args: undefined },
    { command: "open_browser", args: { url: "https://racore.xyz" } },
    { command: "open_external", args: { url: "https://racore.xyz/docs" } },
  ]);
});

test("daemon exit listener forwards typed payload and returns unsubscribe", async () => {
  const mock = mockInterop();
  const bridge = createDesktopBridge(mock.interop);
  const received = [];
  const unsubscribe = await bridge.onDaemonExit((payload) => received.push(payload));

  mock.emit({ code: 7, success: false });

  assert.equal(unsubscribe, mock.unsubscribe);
  assert.deepEqual(received, [{ code: 7, success: false }]);
  assert.deepEqual(mock.calls, [{ event: "racore://daemon-exit" }]);
});

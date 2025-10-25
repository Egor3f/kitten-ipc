import { ParentIPC, ChildIPC } from "kitten-ipc";
export default class IpcApi {
  private ipc: ParentIPC | ChildIPC;

  constructor(ipc: ParentIPC | ChildIPC) {
    this.ipc = ipc;
  }

  async Div(a: number, b: number): Promise<number> {
    const results = await this.ipc.call("Div", a, b);
    return results[0] as number;
  }
}

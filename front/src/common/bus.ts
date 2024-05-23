import mitt from 'mitt';
import type { Emitter } from 'mitt';

type Events = {
  foo: string;
  bar?: number;
  [key: string]: any;
};

interface IBus {
  $on: Emitter<Events>['on'];
  $off: Emitter<Events>['off'];
  $emit: Emitter<Events>['emit'];
}

const emitter = mitt<Events>();
const bus: IBus = {
  $on: emitter.on,
  $off: emitter.off,
  $emit: emitter.emit,
};

export default bus;

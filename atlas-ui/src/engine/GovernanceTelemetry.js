export const GovernanceTelemetry = {
  events: [],
  listeners: new Set(),

  emit(event) {
    this.events.push(event);
    this.listeners.forEach(fn => fn(event));
  },

  subscribe(fn) {
    this.listeners.add(fn);
    return () => this.listeners.delete(fn);
  }
};

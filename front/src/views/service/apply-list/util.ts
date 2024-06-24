export const reverseObj = (originalMap: Object) => {
  Object.fromEntries(Object.entries(originalMap).map(([key, value]) => [value, key]));
};

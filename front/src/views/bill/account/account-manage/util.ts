export const convertToIdNameMap = (arr: Array<{
  id: string,
  name: string,
}>) => {
  return arr.reduce((acc, item) => {
    acc[item.id] = item.name;
    return acc;
  }, {});
}
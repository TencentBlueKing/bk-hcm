export const convertToIdNameMap = (
  arr: Array<{
    id: string;
    name: string;
  }>,
): Record<string, string> => {
  return arr.reduce((acc, item) => {
    acc[item.id] = item.name;
    return acc;
  }, {} as Record<string, string>);
};

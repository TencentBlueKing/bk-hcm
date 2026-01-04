/* eslint-disable */
import { saveAs } from 'file-saver';
import * as XLSX from 'xlsx';

interface CellRange {
  s: { r: number; c: number };
  e: { r: number; c: number };
}

function generateArray(table: HTMLTableElement): [any[][], CellRange[]] {
  var out = [];
  var rows = table.querySelectorAll('tr');
  var ranges = [];
  for (var R = 0; R < rows.length; ++R) {
    var outRow = [];
    var row = rows[R];
    var columns = row.querySelectorAll('td');
    for (var C = 0; C < columns.length; ++C) {
      var cell = columns[C];
      var colspan = cell.getAttribute('colspan');
      var rowspan = cell.getAttribute('rowspan');
      var cellValue: string | number = cell.innerText;
      // 如果是有效数字字符串，转换为数字类型
      if (cellValue !== '' && !isNaN(Number(cellValue))) cellValue = Number(cellValue);

      //Skip ranges
      ranges.forEach(function (range) {
        if (R >= range.s.r && R <= range.e.r && outRow.length >= range.s.c && outRow.length <= range.e.c) {
          for (var i = 0; i <= range.e.c - range.s.c; ++i) outRow.push(null);
        }
      });

      //Handle Row Span
      const colspanNum = Number(colspan) || 1;
      if (rowspan || colspan) {
        const rowspanNum = Number(rowspan) || 1;
        ranges.push({
          s: {
            r: R,
            c: outRow.length,
          },
          e: {
            r: R + rowspanNum - 1,
            c: outRow.length + colspanNum - 1,
          },
        });
      }

      //Handle Value
      outRow.push(cellValue !== '' ? cellValue : null);

      //Handle Colspan
      if (colspanNum > 1) for (var k = 0; k < colspanNum - 1; ++k) outRow.push(null);
    }
    out.push(outRow);
  }
  return [out, ranges];
}

function datenum(v: Date, date1904?: boolean) {
  let serial = (v.getTime() - new Date(Date.UTC(1899, 11, 30)).getTime()) / (24 * 60 * 60 * 1000);
  if (date1904) serial -= 1462;
  return serial;
}

function sheet_from_array_of_arrays(data: any[][], opts?: any) {
  var ws: { [key: string]: any } = {};
  var range = {
    s: {
      c: 10000000,
      r: 10000000,
    },
    e: {
      c: 0,
      r: 0,
    },
  };
  for (var R = 0; R != data.length; ++R) {
    for (var C = 0; C != data[R].length; ++C) {
      if (range.s.r > R) range.s.r = R;
      if (range.s.c > C) range.s.c = C;
      if (range.e.r < R) range.e.r = R;
      if (range.e.c < C) range.e.c = C;
      var cell: { v: any; t?: string; z?: string } = {
        v: data[R][C],
      };
      if (cell.v == null) continue;
      var cell_ref = XLSX.utils.encode_cell({
        c: C,
        r: R,
      });

      if (typeof cell.v === 'number') cell.t = 'n';
      else if (typeof cell.v === 'boolean') cell.t = 'b';
      else if (cell.v instanceof Date) {
        cell.t = 'n';
        cell.z = XLSX.SSF._table[14];
        cell.v = datenum(cell.v);
      } else cell.t = 's';

      ws[cell_ref] = cell;
    }
  }
  if (range.s.c < 10000000) ws['!ref'] = XLSX.utils.encode_range(range);
  return ws;
}

class Workbook {
  SheetNames: string[] = [];
  Sheets: { [key: string]: any } = {};
}

function s2ab(s: string) {
  var buf = new ArrayBuffer(s.length);
  var view = new Uint8Array(buf);
  for (var i = 0; i != s.length; ++i) view[i] = s.charCodeAt(i) & 0xff;
  return buf;
}

export function export_table_to_excel(id: string) {
  var theTable = document.getElementById(id) as HTMLTableElement | null;
  if (!theTable) {
    console.error(`Table element with id "${id}" not found`);
    return;
  }
  var oo = generateArray(theTable);
  var ranges = oo[1];

  /* original data */
  var data = oo[0];
  var ws_name = 'SheetJS';

  var wb = new Workbook(),
    ws = sheet_from_array_of_arrays(data);

  /* add ranges to worksheet */
  // ws['!cols'] = ['apple', 'banan'];
  ws['!merges'] = ranges;

  /* add worksheet to workbook */
  wb.SheetNames.push(ws_name);
  wb.Sheets[ws_name] = ws;

  var wbout = XLSX.write(wb, {
    bookType: 'xlsx',
    bookSST: false,
    type: 'binary',
  });

  saveAs(
    new Blob([s2ab(wbout)], {
      type: 'application/octet-stream',
    }),
    'test.xlsx',
  );
}
export interface ExportOptions {
  header: string[] | string[][];
  data: any[][];
  filename: string;
  autoWidth?: boolean;
  mutipleHeader?: boolean;
  merges?: any[];
  headerIndex?: number;
  maxRowsPerFile?: number;
}

// 单个文件导出的核心逻辑
function exportSingleExcel({
  header,
  data,
  filename,
  autoWidth = true,
  mutipleHeader = false,
  merges = [],
  headerIndex = 0,
}: Omit<ExportOptions, 'maxRowsPerFile'>) {
  data = [...data];
  if (mutipleHeader) {
    // 多行表头时，header 应为 string[][]
    const multiHeader = header as string[][];
    const reversedHeader = [...multiHeader].reverse();
    reversedHeader.forEach((hd) => {
      data.unshift(hd);
    });
  } else {
    data.unshift(header as string[]);
  }
  var ws_name = 'SheetJS';
  var wb = new Workbook(),
    ws = sheet_from_array_of_arrays(data);

  if (autoWidth) {
    /*设置worksheet每列的最大宽度*/
    const colWidth = data.map((row) =>
      row.map((val) => {
        /*先判断是否为null/undefined*/
        if (val == null) {
          return {
            wch: 10,
          };
        } else if (val.toString().charCodeAt(0) > 255) {
          /*再判断是否为中文*/
          return {
            wch: val.toString().length * 2,
          };
        } else {
          return {
            wch: val.toString().length,
          };
        }
      }),
    );
    /*以第一行为初始值*/
    let result = colWidth[headerIndex];
    for (let i = 1; i < colWidth.length; i++) {
      for (let j = 0; j < colWidth[i].length; j++) {
        if (!result[j]) {
          result[j] = { wch: 10 };
        }

        if (result[j]['wch'] < colWidth[i][j]['wch']) {
          result[j]['wch'] = colWidth[i][j]['wch'];
        }
      }
    }
    ws['!cols'] = result;
  }

  if (merges.length) {
    ws['!merges'] = merges;
  }

  /* add worksheet to workbook */
  wb.SheetNames.push(ws_name);
  wb.Sheets[ws_name] = ws;

  var wbout = XLSX.write(wb, {
    bookType: 'xlsx',
    bookSST: false,
    type: 'array', // 使用 array 类型避免 s2ab 转换问题
  });
  saveAs(
    new Blob([wbout], {
      type: 'application/octet-stream',
    }),
    filename + '.xlsx',
  );
}

// 每批最大行数（5万行，避免内存溢出）
export const MAX_ROWS_PER_FILE = 150000;

export function export_json_to_excel(opts: ExportOptions = { header: [], data: [], filename: '' }): Promise<void> {
  let {
    header,
    data,
    filename,
    autoWidth = true,
    mutipleHeader = false, // 是否包含多行 header
    merges = [], // 合并单元格选项
    headerIndex = 0, //以 header 第几列为基础
    maxRowsPerFile = MAX_ROWS_PER_FILE,
  } = opts;
  return new Promise((resolve) => {
    /* original data */
    filename = filename || 'excel-list';

    // 如果数据量不大，直接导出
    if (data.length <= maxRowsPerFile) {
      exportSingleExcel({
        header,
        data,
        filename,
        autoWidth,
        mutipleHeader,
        merges,
        headerIndex,
      });
      resolve();
      return;
    }

    // 数据量大，分批导出多个文件
    const totalParts = Math.ceil(data.length / maxRowsPerFile);
    console.log(`数据量较大（${data.length} 行），将分 ${totalParts} 个文件导出`);

    let completedParts = 0;

    for (let i = 0; i < totalParts; i++) {
      const start = i * maxRowsPerFile;
      const end = Math.min(start + maxRowsPerFile, data.length);
      const chunk = data.slice(start, end);

      // 延迟导出，避免同时创建多个大文件导致内存问题
      setTimeout(() => {
        exportSingleExcel({
          header,
          data: chunk,
          filename: `${filename}_第${i + 1}部分_共${totalParts}部分`,
          autoWidth,
          mutipleHeader,
          merges: i === 0 ? merges : [], // 只有第一个文件保留合并单元格
          headerIndex,
        });
        completedParts++;
        // 所有部分导出完成后 resolve
        if (completedParts === totalParts) {
          resolve();
        }
      }, i * 500); // 每个文件间隔 500ms
    }
  });
}

export interface IComparePickerModel {
  compareType: 'yoy' | 'mom';
  daterange: Date[];
}

export interface IDetailComponentProps {
  currentDate: Date;
  compareDate: Date;
}

export interface IChartCompareProps {
  option: IComparePickerModel;
}

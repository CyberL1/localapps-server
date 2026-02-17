export interface Props {
  id?: string;
  items: Item[];
}

export interface Item {
  title: string;
  onclick?: () => void;
}

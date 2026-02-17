export interface Props {
  id?: string;
  data: object;
  items: Item[];
}

interface Item {
  title: string;
  onclick?: () => void;
}

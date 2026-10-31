/** Route is one matatu route. Some routes have no number painted on, so number and fare are optional. */
export interface Route {
  number?: string;
  operator: string;
  boardsAt: string;
  goesTo: string;
  fareKsh?: number;
}

/** ROUTES were checked on nishukishe.com in October 2026. Routes change, so confirm with the conductor. */
export const ROUTES: Route[] = [
  { number: "46K", operator: "Citi Hoppa", boardsAt: "Kencom / Ambassadeur", goesTo: "Kawangware", fareKsh: 100 },
  { number: "111", operator: "MCA Sacco", boardsAt: "Railways", goesTo: "Ngong", fareKsh: 100 },
  { number: "125", operator: "City Shuttle", boardsAt: "Railways", goesTo: "Ongata Rongai", fareKsh: 100 },
  { number: "45", operator: "Nicco Travellers", boardsAt: "Ronald Ngala Street", goesTo: "Githurai", fareKsh: 100 },
  { number: "58", operator: "Buruburu 58 Travellers Sacco", boardsAt: "Ambassadeur", goesTo: "Buruburu", fareKsh: 150 },
  { operator: "Highrise Kibera Sacco", boardsAt: "Railways", goesTo: "Kibera" },
];

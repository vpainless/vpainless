import { Button } from "@/components/ui/button";
import Link from "next/link"

const asciiart = `
XXXXX           XXXXX
 XXXXX         XXXXX 
    XXX       XXX    
     XXX     XXX     
      XXX   XXX      
       XXX XXX       
        XXXXX        
         XXX         
          X  PainLess
`;

export default function Home() {
	return (
		<div className="grid grid-rows-[20px_1fr_20px] items-center justify-items-center min-h-screen p-8 pb-20 gap-16 sm:p-20 font-[family-name:var(--font-geist-mono)]">
			<main className="flex flex-col gap-[32px] row-start-2 items-center ">
				<pre className="text-gray-400 text-sm sm:text-base font-mono whitespace-pre text-center mb-8">
					{asciiart}
				</pre>
				<pre className="text-black-700 text-sm sm:text-base font-mono whitespace-pre text-center ">
					VPN Creation Made Easy!
				</pre>
				<div className="flex gap-4 items-center flex-col sm:flex-row">
					<Button asChild>
						<Link href="/login">Portal</Link>
					</Button>
				</div>
			</main>
		</div>
	);
}



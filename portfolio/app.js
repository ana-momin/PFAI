const hosted={chat:"https://pfai-signal-chat.onrender.com",auth:"https://pfai-keyline.vercel.app",crawler:"https://pfai-orbit.vercel.app"};
document.querySelectorAll("[data-app]").forEach(link=>link.href=hosted[link.dataset.app]);
const observer=new IntersectionObserver(entries=>entries.forEach(entry=>{if(entry.isIntersecting){entry.target.classList.add("visible");observer.unobserve(entry.target)}}),{threshold:.12});
document.querySelectorAll(".reveal").forEach(el=>observer.observe(el));
const progress=document.querySelector(".progress");
addEventListener("scroll",()=>{const max=document.documentElement.scrollHeight-innerHeight;progress.style.width=`${max?scrollY/max*100:0}%`},{passive:true});
